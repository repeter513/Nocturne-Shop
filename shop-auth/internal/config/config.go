// Application configuration loaded from environment variables.
// Конфигурация приложения, загружаемая из переменных окружения.
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

// Config holds runtime settings for the auth service.
// Config содержит параметры работы сервиса auth.
type Config struct {
	// GRPCPort is the TCP port for the gRPC listener (env: GRPC_PORT, e.g. 8081).
	// GRPCPort — TCP-порт gRPC-слушателя (env: GRPC_PORT, например 8081).
	GRPCPort int16
	// DatabaseURL is the PostgreSQL DSN for auth_db (env: DATABASE_URL).
	// DatabaseURL — DSN PostgreSQL для auth_db (env: DATABASE_URL).
	DatabaseURL string
	// JWTPrivateKey signs access and refresh tokens; loaded from JWT_PRIVATE_KEY_PATH PEM file.
	// JWTPrivateKey подписывает access и refresh токены; загружается из PEM-файла JWT_PRIVATE_KEY_PATH.
	JWTPrivateKey ed25519.PrivateKey
	// JWTPublicKey verifies tokens issued by this or peer services (env: JWT_PUBLIC_KEY_PATH).
	// JWTPublicKey проверяет токены, выданные этим или другими сервисами (env: JWT_PUBLIC_KEY_PATH).
	JWTPublicKey ed25519.PublicKey
	// JWTAccessTTL is access token lifetime (env: JWT_ACCESS_TTL, e.g. "15m").
	// JWTAccessTTL — время жизни access-токена (env: JWT_ACCESS_TTL, например "15m").
	JWTAccessTTL time.Duration
	// JWTRefreshTTL is refresh token lifetime (env: JWT_REFRESH_TTL, e.g. "168h").
	// JWTRefreshTTL — время жизни refresh-токена (env: JWT_REFRESH_TTL, например "168h").
	JWTRefreshTTL time.Duration
	// LogLevel controls structured log verbosity (env: LOG_LEVEL, required, e.g. "info").
	// LogLevel задаёт уровень логирования (env: LOG_LEVEL, обязателен, например "info").
	LogLevel string
}

// Load reads .env and required environment variables.
// Load читает .env и обязательные переменные окружения.
func Load() (*Config, error) {
	// Optional .env in working directory; missing file is OK, other read errors fail.
	// Опциональный .env в рабочей директории; отсутствие файла допустимо, другие ошибки чтения — нет.
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	accessTTL, err := requiredDuration("JWT_ACCESS_TTL")
	if err != nil {
		return nil, err
	}
	refreshTTL, err := requiredDuration("JWT_REFRESH_TTL")
	if err != nil {
		return nil, err
	}

	priv, err := loadPrivateKey(os.Getenv("JWT_PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, err
	}
	pub, err := loadPublicKey(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		return nil, err
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
		return nil, errors.New("LOG_LEVEL environment variable not set")
	}

	return &Config{
		GRPCPort:      int16(port),
		DatabaseURL:   databaseURL,
		JWTPrivateKey: priv,
		JWTPublicKey:  pub,
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
		LogLevel:      logLevel,
	}, nil
}

// GRPCAddr returns the listen address for the gRPC server.
// GRPCAddr возвращает адрес прослушивания gRPC-сервера.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

// requiredDuration parses a Go duration string from env; empty value returns error.
// requiredDuration парсит строку duration из env; пустое значение — ошибка.
func requiredDuration(key string) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, fmt.Errorf("%s environment variable not set", key)
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}

// requiredInt parses a decimal integer from env; empty or invalid value returns error.
// requiredInt парсит целое из env; пустое или невалидное значение — ошибка.
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

// loadPrivateKey reads Ed25519 private key PEM from path; path must be non-empty.
// loadPrivateKey читает PEM приватного Ed25519 ключа; путь обязателен.
func loadPrivateKey(path string) (ed25519.PrivateKey, error) {
	if path == "" {
		return nil, errors.New("JWT_PRIVATE_KEY_PATH required")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH: %w", err)
	}
	return pkgauth.LoadPrivateKeyPEM(b)
}

// loadPublicKey reads Ed25519 public key PEM from path; path must be non-empty.
// loadPublicKey читает PEM публичного Ed25519 ключа; путь обязателен.
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

// requiredURL validates env value as URL with scheme and host (postgres://user:pass@host/db).
// requiredURL проверяет env как URL со scheme и host (postgres://user:pass@host/db).
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
