// Package config loads order service settings from environment variables.
// Пакет config загружает настройки сервиса заказов из переменных окружения.
package config

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
)

// Config holds runtime configuration for the order service.
// Config хранит конфигурацию времени выполнения сервиса заказов.
type Config struct {
	// GRPCPort is the TCP port the gRPC server listens on.
	// GRPCPort — TCP-порт, на котором слушает gRPC-сервер.
	GRPCPort int16
	// DatabaseURL is the PostgreSQL connection string (DSN).
	// DatabaseURL — строка подключения к PostgreSQL (DSN).
	DatabaseURL string
	// LogLevel controls application logging verbosity (default: info).
	// LogLevel управляет детализацией логов приложения (по умолчанию: info).
	LogLevel string
	// CartAddr is the host:port of the cart gRPC service.
	// CartAddr — host:port gRPC-сервиса корзины.
	CartAddr string
	// CatalogAddr is the host:port of the catalog gRPC service.
	// CatalogAddr — host:port gRPC-сервиса каталога.
	CatalogAddr string
	// PaymentAddr is the host:port of the payment gRPC service.
	// PaymentAddr — host:port gRPC-сервиса платежей.
	PaymentAddr string
	// JWTPublicKey is the Ed25519 public key used to verify incoming JWT tokens.
	// JWTPublicKey — публичный ключ Ed25519 для проверки входящих JWT-токенов.
	JWTPublicKey ed25519.PublicKey
}

// Load reads and validates configuration from .env and environment.
// Load читает и проверяет конфигурацию из .env и окружения.
func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	port, err := requiredInt("GRPC_PORT")
	if err != nil {
		return nil, err
	}
	databaseURL, err := requiredURL("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	pub, err := loadPublicKey(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	cartAddr, err := requiredString("CART_GRPC_ADDR")
	if err != nil {
		return nil, err
	}
	catalogAddr, err := requiredString("CATALOG_GRPC_ADDR")
	if err != nil {
		return nil, err
	}
	paymentAddr, err := requiredString("PAYMENT_GRPC_ADDR")
	if err != nil {
		return nil, err
	}

	return &Config{
		GRPCPort:     int16(port),
		DatabaseURL:  databaseURL,
		LogLevel:     logLevel,
		CartAddr:     cartAddr,
		CatalogAddr:  catalogAddr,
		PaymentAddr:  paymentAddr,
		JWTPublicKey: pub,
	}, nil
}

// GRPCAddr returns the listen address for the gRPC server.
// GRPCAddr возвращает адрес прослушивания gRPC-сервера.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

// CartGRPCAddr returns the cart service gRPC address.
// CartGRPCAddr возвращает gRPC-адрес сервиса корзины.
func (c *Config) CartGRPCAddr() string { return c.CartAddr }

// CatalogGRPCAddr returns the catalog service gRPC address.
// CatalogGRPCAddr возвращает gRPC-адрес сервиса каталога.
func (c *Config) CatalogGRPCAddr() string { return c.CatalogAddr }

// PaymentGRPCAddr returns the payment service gRPC address.
// PaymentGRPCAddr возвращает gRPC-адрес сервиса платежей.
func (c *Config) PaymentGRPCAddr() string { return c.PaymentAddr }

// requiredInt reads a required integer environment variable.
// requiredInt читает обязательную целочисленную переменную окружения.
func requiredInt(key string) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("%s environment variable not set", key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return v, nil
}

// requiredURL reads and validates a required database URL.
// requiredURL читает и проверяет обязательный URL базы данных.
func requiredURL(key string) (string, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return "", fmt.Errorf("%s environment variable not set", key)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%s: %w", key, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New(key + ": invalid database url")
	}
	return u.String(), nil
}

// loadPublicKey loads an Ed25519 public key from a PEM file path.
// loadPublicKey загружает публичный ключ Ed25519 из PEM-файла.
func loadPublicKey(path string) (ed25519.PublicKey, error) {
	if path == "" {
		return nil, errors.New("JWT_PUBLIC_KEY_PATH required")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY_PATH: %w", err)
	}
	return pkgauth.LoadPublicKeyPEM(b)
}

// requiredString reads a required string environment variable.
// requiredString читает обязательную строковую переменную окружения.
func requiredString(key string) (string, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return "", fmt.Errorf("%s environment variable not set", key)
	}
	return raw, nil
}
