// Package postgres provides PostgreSQL connection pooling for the payment service.
// Пакет postgres предоставляет пул подключений PostgreSQL для сервиса платежей.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool.
// DB оборачивает пул подключений pgx.
type DB struct {
	// pool is the underlying pgxpool.Pool managing PostgreSQL connections.
	// pool — базовый pgxpool.Pool, управляющий соединениями с PostgreSQL.
	pool *pgxpool.Pool
}

// New opens a PostgreSQL connection pool from the given DSN.
// New открывает пул подключений PostgreSQL по указанному DSN.
func New(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close shuts down the connection pool.
// Close закрывает пул подключений.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying pgx pool.
// Pool возвращает базовый пул pgx.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Ping checks database connectivity.
// Ping проверяет доступность базы данных.
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// WithTx runs fn inside a transaction, rolling back on error.
// WithTx выполняет fn внутри транзакции, откатывая при ошибке.
func (db *DB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
