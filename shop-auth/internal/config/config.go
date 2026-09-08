package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort      int16
	DatabaseURL   string
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
	LogLevel      string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	raw := os.Getenv("JWT_ACCESS_TTL")
	if raw == "" {
		return nil, errors.New("JWT_ACCESS_TTL environment variable not set")
	}

	accessTTL, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}

	raw = os.Getenv("JWT_REFRESH_TTL")
	if raw == "" {
		return nil, errors.New("JWT_REFRESH_TTL environment variable not set")
	}

	refreshTTL, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}

	raw = os.Getenv("JWT_SECRET")
	if raw == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}

	secret, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("JWT_SECRET: %w", err)
	}

	raw = os.Getenv("GRPC_PORT")
	if raw == "" {
		return nil, errors.New("GRPC_PORT environment variable not set")
	}

	port, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("GRPC_PORT: %w", err)
	}

	raw = os.Getenv("DATABASE_URL")
	if raw == "" {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	databaseURL, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL: %w", err)
	}

	LogLevel := os.Getenv("LOG_LEVEL")
	if LogLevel == "" {
		return nil, errors.New("LOG_LEVEL environment variable not set")
	}

	cfg := &Config{
		GRPCPort:      int16(port),
		DatabaseURL:   databaseURL.String(),
		JWTSecret:     string(secret),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
		LogLevel:      LogLevel,
	}
	return cfg, nil
}

func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}
