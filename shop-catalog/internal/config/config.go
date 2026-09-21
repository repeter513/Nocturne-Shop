// Package config loads catalog service settings from environment variables.
// Пакет config загружает настройки сервиса каталога из переменных окружения.
package config

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
)

// Config holds runtime settings for the catalog gRPC server.
// Config содержит runtime-настройки gRPC-сервера каталога.
type Config struct {
	// GRPCPort is the TCP port for gRPC listener (env: GRPC_PORT, e.g. 8082).
	// GRPCPort — TCP-порт gRPC-слушателя (env: GRPC_PORT, например 8082).
	GRPCPort int16
	// DatabaseURL is PostgreSQL DSN for catalog_db (env: DATABASE_URL).
	// DatabaseURL — DSN PostgreSQL для catalog_db (env: DATABASE_URL).
	DatabaseURL string
	// LogLevel controls log verbosity (env: LOG_LEVEL; default "info" if unset).
	// LogLevel задаёт уровень логирования (env: LOG_LEVEL; дефолт "info" если не задан).
	LogLevel string
	// ReservationTTL is default hold duration for stock reservations (env: RESERVATION_TTL; default "5m").
	// ReservationTTL — дефолтная длительность резерва остатков (env: RESERVATION_TTL; дефолт "5m").
	ReservationTTL time.Duration
	// CleanupInterval is how often background job marks expired reservations (env: CLEANUP_INTERVAL; default "1m").
	// CleanupInterval — интервал фоновой пометки просроченных резервов (env: CLEANUP_INTERVAL; дефолт "1m").
	CleanupInterval time.Duration
	// JWTPublicKey verifies Bearer tokens from BFF/order services (env: JWT_PUBLIC_KEY_PATH).
	// JWTPublicKey проверяет Bearer-токены от BFF/order (env: JWT_PUBLIC_KEY_PATH).
	JWTPublicKey ed25519.PublicKey
}

// Load reads configuration from .env and environment variables.
// Load читает конфигурацию из .env и переменных окружения.
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

	// Pagination/reservation defaults: RESERVATION_TTL default 5m, CLEANUP_INTERVAL default 1m.
	// Дефолты резервирования: RESERVATION_TTL 5m, CLEANUP_INTERVAL 1m.
	ttlRaw := os.Getenv("RESERVATION_TTL")
	if ttlRaw == "" {
		ttlRaw = "5m"
	}
	ttl, err := time.ParseDuration(ttlRaw)
	if err != nil {
		return nil, fmt.Errorf("RESERVATION_TTL: %w", err)
	}

	cleanupRaw := os.Getenv("CLEANUP_INTERVAL")
	if cleanupRaw == "" {
		cleanupRaw = "1m"
	}
	cleanup, err := time.ParseDuration(cleanupRaw)
	if err != nil {
		return nil, fmt.Errorf("CLEANUP_INTERVAL: %w", err)
	}

	return &Config{
		GRPCPort:        int16(port),
		DatabaseURL:     databaseURL,
		LogLevel:        logLevel,
		ReservationTTL:  ttl,
		CleanupInterval: cleanup,
		JWTPublicKey:    pub,
	}, nil
}

// GRPCAddr returns the listen address for the gRPC server.
// GRPCAddr возвращает адрес прослушивания gRPC-сервера.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

// ReservationTTLSeconds returns the default reservation TTL in seconds for gRPC handler.
// ReservationTTLSeconds возвращает TTL резерва по умолчанию в секундах для gRPC handler.
func (c *Config) ReservationTTLSeconds() int32 {
	return int32(c.ReservationTTL / time.Second)
}

// requiredInt reads and parses a required integer environment variable.
// requiredInt читает и парсит обязательную целочисленную переменную окружения.
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

// loadPublicKey reads an Ed25519 public key from a PEM file.
// loadPublicKey читает публичный Ed25519 ключ из PEM-файла.
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

// requiredURL reads and validates a required database connection URL.
// requiredURL читает и валидирует обязательный URL подключения к БД.
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
