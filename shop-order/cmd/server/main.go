package main

import (
	"context"
	"github.com/repeter513/shop-order/internal/adapter/payment"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/repeter513/shop-order/internal/adapter/cart"
	"github.com/repeter513/shop-order/internal/adapter/catalog"
	"github.com/repeter513/shop-order/internal/adapter/postgres"
	"github.com/repeter513/shop-order/internal/config"
	ordergrpc "github.com/repeter513/shop-order/internal/grpc"
	"github.com/repeter513/shop-order/internal/service"
	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cartClient, err := cart.New(ctx, cfg.CartGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer cartClient.Close()

	catalogClient, err := catalog.New(ctx, cfg.CatalogGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer catalogClient.Close()

	paymentClient, err := payment.New(ctx, cfg.PaymentGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer paymentClient.Close()

	orderRepo := postgres.NewOrderRepo(db)
	orderSvc := service.NewOrderService(orderRepo, cartClient, catalogClient, paymentClient)
	handler := ordergrpc.NewHandler(orderSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(cfg.JWTSecret)))
	orderv1.RegisterOrderServiceServer(srv, handler)
	reflection.Register(srv)

	go func() {
		log.Println("gRPC listen", cfg.GRPCAddr())
		if err := srv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	srv.GracefulStop()
}
