// Package config loads cart service configuration from environment variables.
// Пакет config загружает конфигурацию сервиса корзины из переменных окружения.
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

// Config holds runtime settings for the cart service.
// Config хранит параметры работы сервиса корзины.
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
	// CatalogAddr is the host:port of the catalog gRPC service.
	// CatalogAddr — host:port gRPC-сервиса каталога.
	CatalogAddr string
	// JWTPublicKey is the Ed25519 public key used to verify incoming JWT tokens.
	// JWTPublicKey — публичный ключ Ed25519 для проверки входящих JWT-токенов.
	JWTPublicKey ed25519.PublicKey
}

// Load reads configuration from .env and environment variables.
// Load читает конфигурацию из .env и переменных окружения.
func Load() (*Config, error) {
	// Step 1: load optional .env file for local development.
	// Шаг 1: загрузка необязательного .env для локальной разработки.
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

	catalogAddr := os.Getenv("CATALOG_GRPC_ADDR")
	if catalogAddr == "" {
		return nil, errors.New("CATALOG_GRPC_ADDR environment variable not set")
	}

	return &Config{
		GRPCPort:     int16(port),
		DatabaseURL:  databaseURL,
		LogLevel:     logLevel,
		CatalogAddr:  catalogAddr,
		JWTPublicKey: pub,
	}, nil
}

// GRPCAddr returns the listen address for the gRPC server.
// GRPCAddr возвращает адрес прослушивания gRPC-сервера.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

// CatalogGRPCAddr returns the catalog service gRPC address.
// CatalogGRPCAddr возвращает gRPC-адрес сервиса каталога.
func (c *Config) CatalogGRPCAddr() string {
	return c.CatalogAddr
}

// requiredInt parses a required integer environment variable.
// requiredInt разбирает обязательную целочисленную переменную окружения.
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

// loadPublicKey reads and parses an Ed25519 public key from a PEM file.
// loadPublicKey читает и разбирает публичный ключ Ed25519 из PEM-файла.
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

// requiredURL parses and validates a required database URL.
// requiredURL разбирает и проверяет обязательный URL базы данных.
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
