// Command server starts the payment gRPC microservice.
// Команда server запускает gRPC-микросервис платежей.
package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/repeter513/shop-payment/internal/adapter/order"
	"github.com/repeter513/shop-payment/internal/adapter/postgres"
	"github.com/repeter513/shop-payment/internal/config"
	paygrpc "github.com/repeter513/shop-payment/internal/grpc"
	"github.com/repeter513/shop-payment/internal/service"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// main wires dependencies and runs the gRPC server until shutdown.
// main собирает зависимости и запускает gRPC-сервер до завершения работы.
func main() {
	// Step 1: load configuration from environment.
	// Шаг 1: загрузка конфигурации из окружения.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Step 2: connect to PostgreSQL for payment persistence.
	// Шаг 2: подключение к PostgreSQL для хранения платежей.
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Step 3: dial order service to fetch order totals for payment amount.
	// Шаг 3: подключение к сервису заказов для получения суммы заказа при оплате.
	orderClient, err := order.New(ctx, cfg.OrderGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer orderClient.Close()

	// Step 4: wire repository → service → gRPC handler.
	// Шаг 4: сборка цепочки repository → service → gRPC handler.
	paymentRepo := postgres.NewPaymentRepo(db)
	paymentSvc := service.NewPaymentService(paymentRepo, orderClient)
	handler := paygrpc.NewHandler(paymentSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	// Step 5: register JWT auth interceptor with payment-specific audience.
	// Шаг 5: регистрация JWT-интерцептора с audience, специфичным для платежей.
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudiencePayment)
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier)))
	paymentv1.RegisterPaymentServiceServer(srv, handler)
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
