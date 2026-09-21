// Package repository defines persistence interfaces for the cart domain.
// Пакет repository определяет интерфейсы хранения для домена корзины.
package repository

import (
	"context"

	"github.com/repeter513/shop-cart/internal/domain"
)

// CartItemRepository persists cart line items per user.
// CartItemRepository сохраняет позиции корзины для каждого пользователя.
type CartItemRepository interface {
	// GetItems returns all cart items for the given user.
	// GetItems возвращает все позиции корзины указанного пользователя.
	GetItems(ctx context.Context, userID int) ([]domain.CartItems, error)
	// UpsertItems inserts or updates a cart item quantity (ON CONFLICT merge).
	// UpsertItems вставляет или обновляет количество позиции (объединение при конфликте).
	UpsertItems(ctx context.Context, userID, productID int, quantity int32) error
	// UpdateItems sets the quantity of an existing cart item.
	// UpdateItems устанавливает количество существующей позиции корзины.
	UpdateItems(ctx context.Context, userID, productID int, quantity int32) error
	// DeleteItems removes a product from the user's cart.
	// DeleteItems удаляет товар из корзины пользователя.
	DeleteItems(ctx context.Context, userID, productID int) error
	// Clear removes all items from the user's cart.
	// Clear удаляет все позиции из корзины пользователя.
	Clear(ctx context.Context, userID int) error
}
