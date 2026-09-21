// Package repository defines stock reservation and atomic stock interfaces.
// Пакет repository определяет интерфейсы резервирования и атомарной работы с остатками.
package repository

import (
	"context"

	"github.com/repeter513/shop-catalog/internal/domain"
)

// ReservationRepository manages stock reservation persistence and queries.
// ReservationRepository управляет сохранением и запросами резервов остатков.
type ReservationRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.StockReservation, error)
	FindByOrderID(ctx context.Context, orderID int64) (*domain.StockReservation, error)
	ExpireOverdue(ctx context.Context) error
	GetActiveReservations(ctx context.Context, productID int64) ([]*domain.StockReservation, error)
}

// AtomicStockManager performs reserve, release, and confirm in database transactions.
// AtomicStockManager выполняет резерв, снятие и подтверждение в транзакциях БД.
// Model A: product.stock is physical; active reservations reduce available only.
// Модель A: product.stock — физический остаток; активные резервы уменьшают только доступное.
type AtomicStockManager interface {
	// ReserveWithTransaction is idempotent per order_id: merges into existing active reservation.
	// ReserveWithTransaction идемпотентен по order_id: объединяет с существующим активным резервом.
	ReserveWithTransaction(ctx context.Context, items map[int64]int32, orderID int64, ttlSeconds int32) (*domain.StockReservation, error)
	ReleaseWithTransaction(ctx context.Context, reservationID int64, items map[int64]int32) error
	ConfirmWithTransaction(ctx context.Context, reservationID int64) error
}
