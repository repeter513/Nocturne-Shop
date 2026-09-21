// Package postgres implements stock reservation persistence.
// Пакет postgres реализует хранение резервов остатков.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-catalog/internal/domain"
)

// ReservationRepo provides PostgreSQL-backed reservation storage.
// ReservationRepo предоставляет хранение резервов через PostgreSQL.
type ReservationRepo struct {
	// db provides pool access and transaction helper.
	// db предоставляет доступ к pool и helper транзакций.
	db *DB
}

// NewReservationRepo creates a ReservationRepo bound to the given DB.
// NewReservationRepo создаёт ReservationRepo, привязанный к указанной БД.
func NewReservationRepo(db *DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

// FindByID loads a reservation by primary key.
// FindByID загружает резерв по первичному ключу.
func (r *ReservationRepo) FindByID(ctx context.Context, id int64) (*domain.StockReservation, error) {
	return r.scanOne(ctx, r.db.Pool(), `SELECT id, order_id, items, status, expires_at, created_at FROM stock_reservations WHERE id = $1`, id)
}

// FindByOrderID loads the reservation for a given order.
// FindByOrderID загружает резерв для указанного заказа.
// order_id is UNIQUE — at most one reservation row per order.
// order_id UNIQUE — не более одной строки резерва на заказ.
func (r *ReservationRepo) FindByOrderID(ctx context.Context, orderID int64) (*domain.StockReservation, error) {
	return r.scanOne(ctx, r.db.Pool(), `SELECT id, order_id, items, status, expires_at, created_at FROM stock_reservations WHERE order_id = $1`, orderID)
}

// ExpireOverdue marks active reservations past expires_at as expired.
// ExpireOverdue помечает активные резервы с истёкшим expires_at как expired.
// Idempotent: re-running only affects rows still active and overdue.
// Идемпотентно: повторный запуск затрагивает только ещё active и просроченные строки.
func (r *ReservationRepo) ExpireOverdue(ctx context.Context) error {
	_, err := r.db.Pool().Exec(ctx, `
		UPDATE stock_reservations SET status = $1
		WHERE status = $2 AND expires_at <= NOW()`,
		domain.ReservationExpired, domain.ReservationActive,
	)
	return err
}

// GetActiveReservations returns non-expired active reservations containing a product.
// GetActiveReservations возвращает непросроченные активные резервы с указанным товаром.
// Uses JSONB ? operator with string product_id key; GIN index idx_reservations_items supports this.
// Использует JSONB ? с ключом product_id; GIN-индекс idx_reservations_items поддерживает это.
func (r *ReservationRepo) GetActiveReservations(ctx context.Context, productID int64) ([]*domain.StockReservation, error) {
	key := strconv.FormatInt(productID, 10)
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, order_id, items, status, expires_at, created_at
		FROM stock_reservations
		WHERE status = $1 AND expires_at > NOW() AND items ? $2`,
		domain.ReservationActive, key,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.StockReservation
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

// sumActiveReservedTx totals reserved quantity for a product within a transaction.
// sumActiveReservedTx суммирует зарезервированное количество товара внутри транзакции.
// Used during reserve merge to compute available = physical - sum(other active reservations).
// Используется при merge резерва: available = physical - sum(другие active резервы).
func (r *ReservationRepo) sumActiveReservedTx(ctx context.Context, tx pgx.Tx, productID int64) (int32, error) {
	key := strconv.FormatInt(productID, 10)
	var sum int32
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM((items->>$1)::int), 0)::int
		FROM stock_reservations
		WHERE status = $2 AND expires_at > NOW() AND items ? $1`,
		key, domain.ReservationActive,
	).Scan(&sum)
	return sum, err
}

// scanOne executes a single-row query and maps the result to a reservation.
// scanOne выполняет однострочный запрос и преобразует результат в резерв.
func (r *ReservationRepo) scanOne(ctx context.Context, q querier, sql string, arg any) (*domain.StockReservation, error) {
	row := q.QueryRow(ctx, sql, arg)
	res, err := scanReservation(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

// querier abstracts QueryRow for pool and transaction contexts.
// querier абстрагирует QueryRow для пула и транзакций.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// scanReservation maps a database row to a domain StockReservation.
// scanReservation преобразует строку БД в доменный StockReservation.
func scanReservation(row rowScanner) (*domain.StockReservation, error) {
	var res domain.StockReservation
	var itemsJSON []byte
	var status string
	if err := row.Scan(&res.ID, &res.OrderID, &itemsJSON, &status, &res.ExpiresAt, &res.CreatedAt); err != nil {
		return nil, err
	}
	res.Status = domain.ReservationStatus(status)
	if err := json.Unmarshal(itemsJSON, &res.Items); err != nil {
		return nil, fmt.Errorf("decode items: %w", err)
	}
	if res.Items == nil {
		res.Items = map[int64]int32{}
	}
	return &res, nil
}

// reservationTTL converts TTL seconds to duration with a 5-minute default.
// reservationTTL преобразует TTL в секундах в duration с дефолтом 5 минут.
func reservationTTL(ttlSeconds int32) time.Duration {
	if ttlSeconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(ttlSeconds) * time.Second
}
