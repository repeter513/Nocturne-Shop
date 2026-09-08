package repository

import (
	"context"

	"github.com/repeter513/shop-catalog/internal/domain"
)

type ReservationRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.StockReservation, error)
	FindByOrderID(ctx context.Context, orderID int64) (*domain.StockReservation, error)
	ExpireOverdue(ctx context.Context) error
	GetActiveReservations(ctx context.Context, productID int64) ([]*domain.StockReservation, error)
}

// AtomicStockManager — model A: product.stock is physical; active reservations reduce available only.
type AtomicStockManager interface {
	ReserveWithTransaction(ctx context.Context, items map[int64]int32, orderID int64, ttlSeconds int32) (*domain.StockReservation, error)
	ReleaseWithTransaction(ctx context.Context, reservationID int64, items map[int64]int32) error
	ConfirmWithTransaction(ctx context.Context, reservationID int64) error
}
