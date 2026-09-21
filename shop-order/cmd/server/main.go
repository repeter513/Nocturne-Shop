// Package main starts the order service gRPC server.
// Пакет main запускает gRPC-сервер сервиса заказов.
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

// main wires dependencies and runs the gRPC server until shutdown.
// main связывает зависимости и запускает gRPC-сервер до завершения работы.
func main() {
	// Step 1: load configuration from environment.
	// Шаг 1: загрузка конфигурации из окружения.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Step 2: connect to PostgreSQL for order persistence.
	// Шаг 2: подключение к PostgreSQL для хранения заказов.
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Step 3: dial downstream services (cart, catalog, payment).
	// Шаг 3: подключение к downstream-сервисам (cart, catalog, payment).
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

	// Step 4: wire repository → service → gRPC handler.
	// Шаг 4: сборка цепочки repository → service → gRPC handler.
	orderRepo := postgres.NewOrderRepo(db)
	orderSvc := service.NewOrderService(orderRepo, cartClient, catalogClient, paymentClient)
	handler := ordergrpc.NewHandler(orderSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	// Step 5: register JWT auth interceptor with order-specific audience.
	// Шаг 5: регистрация JWT-интерцептора с audience, специфичным для заказов.
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceOrder)
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier)))
	orderv1.RegisterOrderServiceServer(srv, handler)
	reflection.Register(srv)

	go func() {
		log.Println("gRPC listen", cfg.GRPCAddr())
		if err := srv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Step 6: graceful shutdown on SIGINT/SIGTERM.
	// Шаг 6: корректное завершение по SIGINT/SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	srv.GracefulStop()
}
