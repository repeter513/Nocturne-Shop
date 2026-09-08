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

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clients, err := client.New(ctx, cfg.AuthAddr, cfg.CatalogAddr, cfg.CartAddr, cfg.OrderAddr, cfg.PaymentAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer clients.Close()

	mux := http.NewServeMux()
	bffhttp.NewHandler(clients).Register(mux, cfg.CORSOrigins)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr(),
		Handler: mux,
	}

	go func() {
		log.Println("HTTP listen", cfg.HTTPAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
