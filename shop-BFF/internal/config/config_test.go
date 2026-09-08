package config_test

import (
	"os"
	"testing"

	"github.com/repeter513/shop-BFF/internal/config"
)

func TestLoadFromEnvFile(t *testing.T) {
	for _, key := range []string{
		"HTTP_PORT", "AUTH_GRPC_ADDR", "CATALOG_GRPC_ADDR",
		"CART_GRPC_ADDR", "ORDER_GRPC_ADDR", "PAYMENT_GRPC_ADDR",
		"LOG_LEVEL", "CORS_ORIGINS",
	} {
		os.Unsetenv(key)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPPort != 8080 {
		t.Fatalf("HTTPPort = %d, want 8080", cfg.HTTPPort)
	}
	if cfg.AuthAddr != "localhost:8081" {
		t.Fatalf("AuthAddr = %q, want localhost:8081", cfg.AuthAddr)
	}
	if cfg.CatalogAddr != "localhost:8082" {
		t.Fatalf("CatalogAddr = %q, want localhost:8082", cfg.CatalogAddr)
	}
	if cfg.CartAddr != "localhost:8083" {
		t.Fatalf("CartAddr = %q, want localhost:8083", cfg.CartAddr)
	}
	if cfg.OrderAddr != "localhost:8084" {
		t.Fatalf("OrderAddr = %q, want localhost:8084", cfg.OrderAddr)
	}
	if cfg.PaymentAddr != "localhost:8086" {
		t.Fatalf("PaymentAddr = %q, want localhost:8086", cfg.PaymentAddr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.CORSOrigins != "*" {
		t.Fatalf("CORSOrigins = %q, want *", cfg.CORSOrigins)
	}
}
