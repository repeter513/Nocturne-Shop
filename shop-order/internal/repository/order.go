package repository

import (
	"context"
	"github.com/repeter513/shop-order/internal/domain"
)

type OrderRepository interface {
	Create(
		ctx context.Context,
		order *domain.Order,
	) error
	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.Order, error)
	ListByUser(
		ctx context.Context,
		userID int64,
		page,
		pageSize int32,
	) ([]domain.Order, int32, error)
	UpdateStatus(
		ctx context.Context,
		id int64,
		status domain.OrderStatus,
		paymentID int64,
	) error
}
