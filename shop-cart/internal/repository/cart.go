package repository

import (
	"context"

	"github.com/repeter513/shop-cart/internal/domain"
)

type CartItemRepository interface {
	GetItems(ctx context.Context, userID int) ([]domain.CartItems, error)
	UpsertItems(ctx context.Context, userID, productID int, quantity int32) error
	UpdateItems(ctx context.Context, userID, productID int, quantity int32) error
	DeleteItems(ctx context.Context, userID, productID int) error
	Clear(ctx context.Context, userID int) error
}
