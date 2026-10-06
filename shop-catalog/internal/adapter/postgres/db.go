// Package postgres implements catalog persistence with PostgreSQL.
// Пакет postgres реализует хранение каталога в PostgreSQL.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool for the catalog service.
// DB оборачивает пул соединений pgx для сервиса каталога.
type DB struct {
	// pool is the shared pgxpool used by all repos in this process.
	// pool — общий pgxpool, используемый всеми repos в процессе.
	pool *pgxpool.Pool
}

// New opens a connection pool and verifies connectivity.
// maxConns/minConns come from config (DB_MAX_CONNS / DB_MIN_CONNS); defaults 10/2 per instance.
// New открывает пул соединений и проверяет подключение.
// maxConns/minConns из config (DB_MAX_CONNS / DB_MIN_CONNS); дефолт 10/2 на инстанс.
func New(ctx context.Context, dsn string, maxConns, minConns int32) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
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
// Close закрывает пул соединений.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying pgx pool for queries.
// Pool возвращает базовый пул pgx для запросов.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Ping checks database connectivity.
// Ping проверяет доступность базы данных.
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// WithTx runs fn inside a transaction, rolling back on error.
// WithTx выполняет fn внутри транзакции с откатом при ошибке.
// defer Rollback is no-op after successful Commit.
// defer Rollback — no-op после успешного Commit.
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
