// Command server starts the cart gRPC microservice.
// Команда server запускает gRPC-микросервис корзины.
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

	// Step 2: connect to PostgreSQL for cart persistence.
	// Шаг 2: подключение к PostgreSQL для хранения корзины.
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Step 3: dial catalog service for product enrichment and stock checks.
	// Шаг 3: подключение к сервису каталога для обогащения товаров и проверки остатков.
	catalogClient, err := catalog.New(ctx, cfg.CatalogGRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	defer catalogClient.Close()

	// Step 4: wire repository → service → gRPC handler.
	// Шаг 4: сборка цепочки repository → service → gRPC handler.
	cartRepo := postgres.NewCartRepo(db)
	cartSvc := service.NewCartService(cartRepo, catalogClient)
	handler := cartgrpc.NewHandler(cartSvc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	// Step 5: register JWT auth interceptor with cart-specific audience.
	// Шаг 5: регистрация JWT-интерцептора с audience, специфичным для корзины.
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceCart)
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier)))
	cartv1.RegisterCartServiceServer(srv, handler)
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
