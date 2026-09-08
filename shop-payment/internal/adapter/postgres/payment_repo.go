package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/repeter513/shop-payment/internal/domain"
	"github.com/repeter513/shop-payment/internal/repository"
)

var _ repository.PaymentRepository = (*PaymentRepo)(nil)

type PaymentRepo struct {
	db *DB
}

func NewPaymentRepo(db *DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	err := r.db.Pool().QueryRow(ctx, `
		INSERT INTO payments (order_id, user_id, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		p.OrderID, p.UserID, p.Amount, p.Status,
	).Scan(&p.ID, &p.CreatedAt)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return r.loadByOrderID(ctx, p)
	}
	return err
}

func (r *PaymentRepo) loadByOrderID(ctx context.Context, p *domain.Payment) error {
	return r.db.Pool().QueryRow(ctx, `
		SELECT id, order_id, user_id, amount, status, created_at
		FROM payments
		WHERE order_id = $1`, p.OrderID,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.CreatedAt)
}

func (r *PaymentRepo) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, order_id, user_id, amount, status, created_at
		FROM payments
		WHERE id = $1`, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) List(
	ctx context.Context,
	userID, orderID int64,
	page, pageSize int32,
) ([]domain.Payment, int32, error) {
	var total int32
	if err := r.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM payments
		WHERE ($1 = 0 OR user_id = $1)
		  AND ($2 = 0 OR order_id = $2)`,
		userID, orderID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, order_id, user_id, amount, status, created_at
		FROM payments
		WHERE ($1 = 0 OR user_id = $1)
		  AND ($2 = 0 OR order_id = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		userID, orderID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}
