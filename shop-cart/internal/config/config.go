package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort    int16
	DatabaseURL string
	LogLevel    string
	CatalogAddr string
	JWTSecret   []byte
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

	raw := os.Getenv("JWT_SECRET")
	if raw == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}

	secret, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("JWT_SECRET: %w", err)
	}

	catalogAddr := os.Getenv("CATALOG_GRPC_ADDR")
	if catalogAddr == "" {
		return nil, errors.New("CATALOG_GRPC_ADDR environment variable not set")
	}

	return &Config{
		GRPCPort:    int16(port),
		DatabaseURL: databaseURL,
		LogLevel:    logLevel,
		CatalogAddr: catalogAddr,
		JWTSecret:   secret,
	}, nil
}

func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

func (c *Config) CatalogGRPCAddr() string {
	return c.CatalogAddr
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
