package main

import (
	"context"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/repeter513/shop-catalog/internal/adapter/postgres"
	"github.com/repeter513/shop-catalog/internal/config"
	cataloggrpc "github.com/repeter513/shop-catalog/internal/grpc"
	"github.com/repeter513/shop-catalog/internal/service"
	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepo := postgres.NewProductRepo(db)
	reservationRepo := postgres.NewReservationRepo(db)
	stockManager := postgres.NewStockManager(db, productRepo, reservationRepo)
	catalogSvc := service.NewCatalogService(productRepo, reservationRepo, stockManager)

	go runReservationCleanup(ctx, catalogSvc, cfg.CleanupInterval)

	handler := cataloggrpc.NewHandler(catalogSvc, cfg.ReservationTTLSeconds())
	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(srv, handler)
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

func runReservationCleanup(ctx context.Context, svc *service.CatalogService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.CleanupExpiredReservations(ctx); err != nil {
				log.Println("cleanup expired reservations:", err)
			}
		}
	}
}
