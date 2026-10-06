// Package main starts the BFF HTTP gateway server.
// Пакет main запускает HTTP-шлюз BFF.
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/repeter513/shop-BFF/internal/client"
	"github.com/repeter513/shop-BFF/internal/config"
	bffhttp "github.com/repeter513/shop-BFF/internal/http"
	"github.com/repeter513/shop-BFF/internal/logx"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
)

// main wires gRPC clients and runs the HTTP server until shutdown.
// main связывает gRPC-клиенты и запускает HTTP-сервер до завершения работы.
func main() {
	// Load env-based config (.env + process environment).
	// Загружаем конфигурацию из .env и переменных окружения.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logx.New(cfg.LogLevel)
	slog.SetDefault(logger)

	// Bounded dial context — fail fast if a backend is unreachable.
	// Ограниченный контекст подключения — быстрый отказ, если backend недоступен.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open insecure gRPC channels to all five microservices.
	// Открываем незащищённые gRPC-каналы ко всем пяти микросервисам.
	clients, err := client.New(ctx, cfg.AuthAddr, cfg.CatalogAddr, cfg.CartAddr, cfg.OrderAddr, cfg.PaymentAddr)
	if err != nil {
		logger.Error("grpc clients", slog.Any("err", err))
		os.Exit(1)
	}
	defer clients.Close()

	// Register REST routes; each handler forwards to the matching gRPC stub.
	// Регистрируем REST-маршруты; каждый обработчик проксирует в соответствующий gRPC-стаб.
	jwt := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceAuth)
	authRL := bffhttp.NewIPRateLimit(cfg.AuthRateLimitPerMin, time.Minute)
	authRefreshRL := bffhttp.NewIPRateLimit(cfg.AuthRefreshRateLimitPerMin, time.Minute)

	mux := http.NewServeMux()
	bffhttp.NewHandler(clients, jwt, authRL, authRefreshRL).Register(mux, cfg.CORSOrigins)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr(),
		Handler: mux,
	}

	// Listen in a goroutine so the main thread can wait for SIGINT/SIGTERM.
	// Слушаем в горутине, чтобы основной поток ждал SIGINT/SIGTERM.
	go func() {
		logger.Info("HTTP listen", slog.String("addr", cfg.HTTPAddr()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP serve", slog.Any("err", err))
		}
	}()

	// Graceful shutdown on Ctrl+C or container stop signal.
	// Корректное завершение по Ctrl+C или сигналу остановки контейнера.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown", slog.Any("err", err))
	}
}
