package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-order/internal/domain"
	"github.com/repeter513/shop-order/internal/repository"
)

var _ repository.OrderRepository = (*OrderRepo)(nil)

type OrderRepo struct {
	db *DB
}

func NewOrderRepo(db *DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		var paymentID *int64
		if order.PaymentID != 0 {
			paymentID = &order.PaymentID
		}

		err := tx.QueryRow(ctx, `
			INSERT INTO orders (user_id, status, total_price, payment_id)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at`,
			order.UserID, order.Status, order.TotalPrice, paymentID,
		).Scan(&order.ID, &order.CreatedAt)
		if err != nil {
			return err
		}

		for _, item := range order.Items {
			_, err := tx.Exec(ctx, `
				INSERT INTO order_items (order_id, product_id, quantity, name, price)
				VALUES ($1, $2, $3, $4, $5)`,
				order.ID, item.ProductID, item.Quantity, item.Name, item.Price,
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepo) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	var order domain.Order
	var paymentID *int64

	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, status, total_price, payment_id, created_at
		FROM orders
		WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &paymentID, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	if paymentID != nil {
		order.PaymentID = *paymentID
	}

	items, err := r.loadItems(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	order.Items = items[id]
	return &order, nil
}

func (r *OrderRepo) ListByUser(
	ctx context.Context,
	userID int64,
	page, pageSize int32,
) ([]domain.Order, int32, error) {
	var total int32
	if err := r.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, status, total_price, payment_id, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.Order
	var ids []int64

	for rows.Next() {
		var order domain.Order
		var paymentID *int64
		if err := rows.Scan(
			&order.ID, &order.UserID, &order.Status,
			&order.TotalPrice, &paymentID, &order.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if paymentID != nil {
			order.PaymentID = *paymentID
		}
		orders = append(orders, order)
		ids = append(ids, order.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(orders) == 0 {
		return orders, total, nil
	}

	itemsByOrder, err := r.loadItems(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
	}
	return orders, total, nil
}

func (r *OrderRepo) UpdateStatus(
	ctx context.Context,
	id int64,
	status domain.OrderStatus,
	paymentID int64,
) error {
	var pid *int64
	if paymentID != 0 {
		pid = &paymentID
	}

	tag, err := r.db.Pool().Exec(ctx, `
		UPDATE orders
		SET status = $2, payment_id = $3
		WHERE id = $1`,
		id, status, pid,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) loadItems(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT order_id, product_id, quantity, name, price
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY product_id`, orderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]domain.OrderItem)
	for rows.Next() {
		var orderID int64
		var item domain.OrderItem
		if err := rows.Scan(&orderID, &item.ProductID, &item.Quantity, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		out[orderID] = append(out[orderID], item)
	}
	return out, rows.Err()
}
