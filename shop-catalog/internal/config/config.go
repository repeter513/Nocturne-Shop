package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort        int16
	DatabaseURL     string
	LogLevel        string
	ReservationTTL  time.Duration
	CleanupInterval time.Duration
}

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
	}, nil
}

func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

func (c *Config) ReservationTTLSeconds() int32 {
	return int32(c.ReservationTTL / time.Second)
}

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
