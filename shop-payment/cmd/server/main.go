package main

import (
	"context"
	"github.com/repeter513/shop-payment/internal/adapter/postgres"
	"github.com/repeter513/shop-payment/internal/config"
	"github.com/repeter513/shop-payment/internal/service"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	paygrpc "github.com/repeter513/shop-payment/internal/grpc"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"

	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	paymentRepo := postgres.NewPaymentRepo(db)
	paymentSvc := service.NewPaymentService(paymentRepo)
	handler := paygrpc.NewHandler(paymentSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(cfg.JWTSecret)))
	paymentv1.RegisterPaymentServiceServer(srv, handler)
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
