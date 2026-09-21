// Package main starts the BFF HTTP gateway server.
// Пакет main запускает HTTP-шлюз BFF.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/repeter513/shop-BFF/internal/client"
	"github.com/repeter513/shop-BFF/internal/config"
	bffhttp "github.com/repeter513/shop-BFF/internal/http"
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

	// Bounded dial context — fail fast if a backend is unreachable.
	// Ограниченный контекст подключения — быстрый отказ, если backend недоступен.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open insecure gRPC channels to all five microservices.
	// Открываем незащищённые gRPC-каналы ко всем пяти микросервисам.
	clients, err := client.New(ctx, cfg.AuthAddr, cfg.CatalogAddr, cfg.CartAddr, cfg.OrderAddr, cfg.PaymentAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer clients.Close()

	// Register REST routes; each handler forwards to the matching gRPC stub.
	// Регистрируем REST-маршруты; каждый обработчик проксирует в соответствующий gRPC-стаб.
	mux := http.NewServeMux()
	bffhttp.NewHandler(clients).Register(mux, cfg.CORSOrigins)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr(),
		Handler: mux,
	}

	// Listen in a goroutine so the main thread can wait for SIGINT/SIGTERM.
	// Слушаем в горутине, чтобы основной поток ждал SIGINT/SIGTERM.
	go func() {
		log.Println("HTTP listen", cfg.HTTPAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown on Ctrl+C or container stop signal.
	// Корректное завершение по Ctrl+C или сигналу остановки контейнера.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
