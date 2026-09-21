// Catalog service entry point: wires dependencies and starts the gRPC server.
// Точка входа сервиса каталога: собирает зависимости и запускает gRPC-сервер.
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
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
)

// main loads config, starts gRPC server, and handles graceful shutdown.
// main загружает конфиг, запускает gRPC-сервер и обрабатывает graceful shutdown.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Step 1: connect to catalog_db with tuned pool settings (see postgres.New).
	// Шаг 1: подключение к catalog_db с настроенным пулом (см. postgres.New).
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Step 2: wire repos → stock manager → catalog service.
	// Шаг 2: сборка repos → stock manager → catalog service.
	productRepo := postgres.NewProductRepo(db)
	reservationRepo := postgres.NewReservationRepo(db)
	stockManager := postgres.NewStockManager(db, productRepo, reservationRepo)
	catalogSvc := service.NewCatalogService(productRepo, reservationRepo, stockManager)

	// Step 3: background goroutine expires overdue reservations (non-blocking cleanup).
	// Шаг 3: фоновая goroutine помечает просроченные резервы (неблокирующая очистка).
	go runReservationCleanup(ctx, catalogSvc, cfg.CleanupInterval)

	handler := cataloggrpc.NewHandler(catalogSvc, cfg.ReservationTTLSeconds())
	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	// Step 4: JWT verifier with catalog audience; PublicCatalog lists unauthenticated RPCs (if any).
	// Шаг 4: JWT verifier с audience catalog; PublicCatalog — RPC без auth (если есть).
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceCatalog)
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier, pkgauth.PublicCatalog...)))
	catalogv1.RegisterCatalogServiceServer(srv, handler)
	reflection.Register(srv)
	go func() {
		log.Println("gRPC listen", cfg.GRPCAddr())
		if err := srv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Step 5: graceful shutdown on SIGINT/SIGTERM — drain in-flight RPCs.
	// Шаг 5: graceful shutdown по SIGINT/SIGTERM — завершение текущих RPC.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	srv.GracefulStop()
}

// runReservationCleanup periodically marks expired stock reservations.
// runReservationCleanup периодически помечает просроченные резервы остатков.
// Idempotent: ExpireOverdue only updates rows where status=active AND expires_at <= NOW().
// Идемпотентно: ExpireOverdue обновляет только строки status=active AND expires_at <= NOW().
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
