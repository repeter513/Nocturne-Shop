// Package repository defines persistence interfaces for orders.
// Пакет repository определяет интерфейсы хранения заказов.
package repository

import (
	"context"
	"github.com/repeter513/shop-order/internal/domain"
)

// OrderRepository persists and queries order aggregates.
// OrderRepository сохраняет и запрашивает агрегаты заказов.
type OrderRepository interface {
	// Create inserts a new order with its items in a single transaction.
	// Create вставляет новый заказ с его позициями в одной транзакции.
	Create(
		ctx context.Context,
		order *domain.Order,
	) error
	// GetByID returns an order by primary key with all line items.
	// GetByID возвращает заказ по первичному ключу со всеми позициями.
	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.Order, error)
	// ListByUser returns paginated orders for a user and total count.
	// ListByUser возвращает постраничный список заказов пользователя и общее количество.
	ListByUser(
		ctx context.Context,
		userID int64,
		page,
		pageSize int32,
	) ([]domain.Order, int32, error)
	// UpdateStatus changes order status and optional payment ID.
	// UpdateStatus изменяет статус заказа и необязательный ID платежа.
	UpdateStatus(
		ctx context.Context,
		id int64,
		status domain.OrderStatus,
		paymentID int64,
	) error
}
