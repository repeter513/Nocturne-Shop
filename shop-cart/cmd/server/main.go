package main

import (
	"context"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/repeter513/shop-cart/internal/adapter/catalog"
	"github.com/repeter513/shop-cart/internal/adapter/postgres"
	"github.com/repeter513/shop-cart/internal/config"
	cartgrpc "github.com/repeter513/shop-cart/internal/grpc"
	"github.com/repeter513/shop-cart/internal/service"
	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
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

	catalogClient, err := catalog.New(ctx, cfg.CatalogGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer catalogClient.Close()

	cartRepo := postgres.NewCartRepo(db)
	cartSvc := service.NewCartService(cartRepo, catalogClient)
	handler := cartgrpc.NewHandler(cartSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(cfg.JWTSecret)))
	cartv1.RegisterCartServiceServer(srv, handler)
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
