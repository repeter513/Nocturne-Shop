// Package repository defines persistence interfaces for the payment domain.
// Пакет repository определяет интерфейсы хранения для домена платежей.
package repository

import (
	"context"

	"github.com/repeter513/shop-payment/internal/domain"
)

// PaymentRepository persists and queries payment records.
// PaymentRepository сохраняет и запрашивает записи платежей.
type PaymentRepository interface {
	// Create inserts a new payment record; idempotent on duplicate order_id.
	// Create вставляет новую запись платежа; идемпотентен при дубликате order_id.
	Create(ctx context.Context, payment *domain.Payment) error
	// GetByID returns a payment by its primary key.
	// GetByID возвращает платёж по первичному ключу.
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	// List returns a paginated, filtered list of payments and the total count.
	// List возвращает постраничный отфильтрованный список платежей и общее количество.
	List(ctx context.Context, userID, orderID int64, page, pageSize int32) ([]domain.Payment, int32, error)
}
