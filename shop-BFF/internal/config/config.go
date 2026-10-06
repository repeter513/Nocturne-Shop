// Package config loads BFF settings from environment variables.
// Пакет config загружает настройки BFF из переменных окружения.
package config

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
)

// Config holds runtime configuration for the BFF gateway.
// Config хранит конфигурацию времени выполнения BFF-шлюза.
type Config struct {
	// HTTPPort is the TCP port the BFF HTTP server listens on (default 8080).
	// HTTPPort — TCP-порт HTTP-сервера BFF (по умолчанию 8080).
	HTTPPort int

	// AuthAddr is the gRPC address of the auth microservice (host:port).
	// AuthAddr — gRPC-адрес микросервиса auth (host:port).
	AuthAddr string

	// CatalogAddr is the gRPC address of the catalog microservice.
	// CatalogAddr — gRPC-адрес микросервиса catalog.
	CatalogAddr string

	// CartAddr is the gRPC address of the cart microservice.
	// CartAddr — gRPC-адрес микросервиса cart.
	CartAddr string

	// OrderAddr is the gRPC address of the order microservice.
	// OrderAddr — gRPC-адрес микросервиса order.
	OrderAddr string

	// PaymentAddr is the gRPC address of the payment microservice.
	// PaymentAddr — gRPC-адрес микросервиса payment.
	PaymentAddr string

	// LogLevel controls structured log verbosity (default "info").
	// LogLevel задаёт уровень детализации логов (по умолчанию "info").
	LogLevel string

	// CORSOrigins is the Access-Control-Allow-Origin value (default "*").
	// CORSOrigins — значение заголовка Access-Control-Allow-Origin (по умолчанию "*").
	CORSOrigins string

	// JWTPublicKey verifies access tokens at the HTTP edge (env: JWT_PUBLIC_KEY_PATH).
	JWTPublicKey ed25519.PublicKey

	// AuthRateLimitPerMin caps register/login requests per client IP per minute (default 30).
	AuthRateLimitPerMin int

	// AuthRefreshRateLimitPerMin caps refresh requests per client IP per minute (default 10).
	AuthRefreshRateLimitPerMin int
}

// Load reads and validates configuration from .env and environment.
// Load читает и проверяет конфигурацию из .env и окружения.
func Load() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	port := 8080
	if raw := os.Getenv("HTTP_PORT"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("HTTP_PORT: %w", err)
		}
		port = v
	}

	cfg := &Config{
		HTTPPort:    port,
		AuthAddr:    os.Getenv("AUTH_GRPC_ADDR"),
		CatalogAddr: os.Getenv("CATALOG_GRPC_ADDR"),
		CartAddr:    os.Getenv("CART_GRPC_ADDR"),
		OrderAddr:   os.Getenv("ORDER_GRPC_ADDR"),
		PaymentAddr: os.Getenv("PAYMENT_GRPC_ADDR"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		CORSOrigins: os.Getenv("CORS_ORIGINS"),
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.CORSOrigins == "" {
		cfg.CORSOrigins = "*"
	}

	pub, err := loadPublicKey(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		return nil, err
	}
	cfg.JWTPublicKey = pub

	cfg.AuthRateLimitPerMin = 30
	if raw := os.Getenv("AUTH_RATE_LIMIT_PER_MIN"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("AUTH_RATE_LIMIT_PER_MIN: %w", err)
		}
		cfg.AuthRateLimitPerMin = v
	}

	cfg.AuthRefreshRateLimitPerMin = 10
	if raw := os.Getenv("AUTH_REFRESH_RATE_LIMIT_PER_MIN"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("AUTH_REFRESH_RATE_LIMIT_PER_MIN: %w", err)
		}
		cfg.AuthRefreshRateLimitPerMin = v
	}

	// Every gRPC backend address is mandatory — BFF cannot start without them.
	// Адрес каждого gRPC-backend обязателен — BFF не стартует без них.
	for _, pair := range []struct {
		name string
		val  string
	}{
		{"AUTH_GRPC_ADDR", cfg.AuthAddr},
		{"CATALOG_GRPC_ADDR", cfg.CatalogAddr},
		{"CART_GRPC_ADDR", cfg.CartAddr},
		{"ORDER_GRPC_ADDR", cfg.OrderAddr},
		{"PAYMENT_GRPC_ADDR", cfg.PaymentAddr},
	} {
		if pair.val == "" {
			return nil, fmt.Errorf("%s environment variable not set", pair.name)
		}
	}

	return cfg, nil
}

// HTTPAddr returns the listen address for the HTTP server.
// HTTPAddr возвращает адрес прослушивания HTTP-сервера.
func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

// loadDotEnv searches upward from cwd for a .env file and loads it.
// loadDotEnv ищет .env вверх от текущей директории и загружает его.
func loadPublicKey(path string) (ed25519.PublicKey, error) {
	if path == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY_PATH environment variable not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY_PATH: %w", err)
	}
	return pkgauth.LoadPublicKeyPEM(b)
}

func loadDotEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			return godotenv.Load(path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}
