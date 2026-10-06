// Package main starts the order service gRPC server.
// Пакет main запускает gRPC-сервер сервиса заказов.
package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/repeter513/shop-order/internal/adapter/cart"
	"github.com/repeter513/shop-order/internal/adapter/catalog"
	"github.com/repeter513/shop-order/internal/adapter/payment"
	"github.com/repeter513/shop-order/internal/adapter/postgres"
	"github.com/repeter513/shop-order/internal/config"
	ordergrpc "github.com/repeter513/shop-order/internal/grpc"
	"github.com/repeter513/shop-order/internal/logx"
	"github.com/repeter513/shop-order/internal/service"
	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	logger := logx.New(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx := context.Background()

	// Step 2: connect to PostgreSQL for order persistence.
	// Шаг 2: подключение к PostgreSQL для хранения заказов.
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()

	// Step 3: dial downstream services (cart, catalog, payment).
	// Шаг 3: подключение к downstream-сервисам (cart, catalog, payment).
	cartClient, err := cart.New(ctx, cfg.CartGRPCAddr())
	if err != nil {
		logger.Error("cart dial", slog.String("addr", cfg.CartGRPCAddr()), slog.Any("err", err))
		os.Exit(1)
	}
	defer cartClient.Close()

	catalogClient, err := catalog.New(ctx, cfg.CatalogGRPCAddr())
	if err != nil {
		logger.Error("catalog dial", slog.String("addr", cfg.CatalogGRPCAddr()), slog.Any("err", err))
		os.Exit(1)
	}
	defer catalogClient.Close()

	paymentClient, err := payment.New(ctx, cfg.PaymentGRPCAddr())
	if err != nil {
		logger.Error("payment dial", slog.String("addr", cfg.PaymentGRPCAddr()), slog.Any("err", err))
		os.Exit(1)
	}
	defer paymentClient.Close()

	// Step 4: wire repository → service → gRPC handler.
	// Шаг 4: сборка цепочки repository → service → gRPC handler.
	orderRepo := postgres.NewOrderRepo(db)
	orderSvc := service.NewOrderService(orderRepo, cartClient, catalogClient, paymentClient)
	handler := ordergrpc.NewHandler(orderSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		logger.Error("listen", slog.String("addr", cfg.GRPCAddr()), slog.Any("err", err))
		os.Exit(1)
	}

	// Step 5: register JWT auth interceptor with order-specific audience.
	// Шаг 5: регистрация JWT-интерцептора с audience, специфичным для заказов.
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceOrder)
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier)))
	orderv1.RegisterOrderServiceServer(srv, handler)
	reflection.Register(srv)

	go func() {
		logger.Info("gRPC listen", slog.String("addr", cfg.GRPCAddr()))
		if err := srv.Serve(lis); err != nil {
			logger.Error("gRPC serve", slog.Any("err", err))
		}
	}()

	// Step 6: graceful shutdown on SIGINT/SIGTERM.
	// Шаг 6: корректное завершение по SIGINT/SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	srv.GracefulStop()
}
