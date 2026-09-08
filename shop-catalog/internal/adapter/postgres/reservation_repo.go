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

type ReservationRepo struct {
	db *DB
}

func NewReservationRepo(db *DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) FindByID(ctx context.Context, id int64) (*domain.StockReservation, error) {
	return r.scanOne(ctx, r.db.Pool(), `SELECT id, order_id, items, status, expires_at, created_at FROM stock_reservations WHERE id = $1`, id)
}

func (r *ReservationRepo) FindByOrderID(ctx context.Context, orderID int64) (*domain.StockReservation, error) {
	return r.scanOne(ctx, r.db.Pool(), `SELECT id, order_id, items, status, expires_at, created_at FROM stock_reservations WHERE order_id = $1`, orderID)
}

func (r *ReservationRepo) ExpireOverdue(ctx context.Context) error {
	_, err := r.db.Pool().Exec(ctx, `
		UPDATE stock_reservations SET status = $1
		WHERE status = $2 AND expires_at <= NOW()`,
		domain.ReservationExpired, domain.ReservationActive,
	)
	return err
}

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

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

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

func reservationTTL(ttlSeconds int32) time.Duration {
	if ttlSeconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(ttlSeconds) * time.Second
}
