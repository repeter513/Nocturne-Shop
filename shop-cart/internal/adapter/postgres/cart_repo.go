package postgres

import (
	"context"

	"github.com/repeter513/shop-cart/internal/domain"
	"github.com/repeter513/shop-cart/internal/repository"
)

var _ repository.CartItemRepository = (*CartRepo)(nil)

type CartRepo struct {
	db *DB
}

func NewCartRepo(db *DB) *CartRepo {
	return &CartRepo{db: db}
}

func (r *CartRepo) GetItems(ctx context.Context, userID int) ([]domain.CartItems, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT product_id, quantity
		FROM cart_items
		WHERE user_id = $1
		ORDER BY product_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItems
	for rows.Next() {
		var item domain.CartItems
		if err := rows.Scan(&item.ProductId, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CartRepo) UpsertItems(ctx context.Context, userID, productID int, quantity int32) error {
	_, err := r.db.Pool().Exec(ctx, `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = EXCLUDED.quantity`,
		userID, productID, quantity)
	return err
}

func (r *CartRepo) UpdateItems(ctx context.Context, userID, productID int, quantity int32) error {
	tag, err := r.db.Pool().Exec(ctx, `
		UPDATE cart_items
		SET quantity = $3
		WHERE user_id = $1 AND product_id = $2`,
		userID, productID, quantity)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepo) DeleteItems(ctx context.Context, userID, productID int) error {
	tag, err := r.db.Pool().Exec(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`,
		userID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepo) Clear(ctx context.Context, userID int) error {
	_, err := r.db.Pool().Exec(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1`, userID)
	return err
}
