// Package postgres implements payment persistence with PostgreSQL.
// Пакет postgres реализует хранение платежей в PostgreSQL.
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

// PaymentRepo is a PostgreSQL-backed PaymentRepository.
// PaymentRepo — реализация PaymentRepository на PostgreSQL.
type PaymentRepo struct {
	// db is the connection pool wrapper for executing SQL queries.
	// db — обёртка пула соединений для выполнения SQL-запросов.
	db *DB
}

// NewPaymentRepo creates a payment repository using the given database handle.
// NewPaymentRepo создаёт репозиторий платежей с указанным подключением к БД.
func NewPaymentRepo(db *DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// Create inserts a new payment; on duplicate order_id loads the existing record.
// Create вставляет новый платёж; при дубликате order_id загружает существующую запись.
//
// Idempotency: unique index on order_id (idx_payments_order_id) triggers PG error 23505;
// the existing payment is loaded and returned instead of failing.
// Идемпотентность: уникальный индекс по order_id (idx_payments_order_id) вызывает ошибку PG 23505;
// существующий платёж загружается и возвращается вместо ошибки.
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
	// Step: detect unique-violation on order_id and load existing row (idempotent retry).
	// Шаг: обнаружение нарушения уникальности по order_id и загрузка существующей строки (идемпотентный повтор).
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return r.loadByOrderID(ctx, p)
	}
	return err
}

// loadByOrderID fetches a payment by order ID into p.
// loadByOrderID загружает платёж по ID заказа в p.
func (r *PaymentRepo) loadByOrderID(ctx context.Context, p *domain.Payment) error {
	return r.db.Pool().QueryRow(ctx, `
		SELECT id, order_id, user_id, amount, status, created_at
		FROM payments
		WHERE order_id = $1`, p.OrderID,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.CreatedAt)
}

// GetByID returns a payment by its primary key.
// GetByID возвращает платёж по первичному ключу.
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

// List returns a paginated, filtered list of payments and the total count.
// List возвращает постраничный отфильтрованный список платежей и общее количество.
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
