package repository

import (
	"context"

	"github.com/repeter513/shop-payment/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	List(ctx context.Context, userID, orderID int64, page, pageSize int32) ([]domain.Payment, int32, error)
}
