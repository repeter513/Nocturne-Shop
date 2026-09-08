package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort     int
	AuthAddr     string
	CatalogAddr  string
	CartAddr     string
	OrderAddr    string
	PaymentAddr  string
	LogLevel     string
	CORSOrigins  string
}

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

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
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
