// Package postgres implements cart persistence with PostgreSQL.
// Пакет postgres реализует хранение корзины в PostgreSQL.
package postgres

import (
	"context"

	"github.com/repeter513/shop-cart/internal/domain"
	"github.com/repeter513/shop-cart/internal/repository"
)

var _ repository.CartItemRepository = (*CartRepo)(nil)

// CartRepo is a PostgreSQL-backed CartItemRepository.
// CartRepo — реализация CartItemRepository на PostgreSQL.
type CartRepo struct {
	// db is the connection pool wrapper for executing SQL queries.
	// db — обёртка пула соединений для выполнения SQL-запросов.
	db *DB
}

// NewCartRepo creates a cart repository using the given database handle.
// NewCartRepo создаёт репозиторий корзины с указанным подключением к БД.
func NewCartRepo(db *DB) *CartRepo {
	return &CartRepo{db: db}
}

// GetItems returns all cart items for the given user.
// GetItems возвращает все позиции корзины указанного пользователя.
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

// UpsertItems inserts or updates a cart item quantity.
// UpsertItems вставляет или обновляет количество позиции в корзине.
func (r *CartRepo) UpsertItems(ctx context.Context, userID, productID int, quantity int32) error {
	// ON CONFLICT (user_id, product_id) merges quantities for the same product line.
	// ON CONFLICT (user_id, product_id) объединяет количества для одной позиции.
	_, err := r.db.Pool().Exec(ctx, `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = EXCLUDED.quantity`,
		userID, productID, quantity)
	return err
}

// UpdateItems sets the quantity of an existing cart item.
// UpdateItems устанавливает количество существующей позиции корзины.
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

// DeleteItems removes a product from the user's cart.
// DeleteItems удаляет товар из корзины пользователя.
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

// Clear removes all items from the user's cart.
// Clear удаляет все позиции из корзины пользователя.
func (r *CartRepo) Clear(ctx context.Context, userID int) error {
	_, err := r.db.Pool().Exec(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1`, userID)
	return err
}
