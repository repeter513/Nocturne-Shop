// Package domain defines order entities and domain errors.
// Пакет domain определяет сущности заказов и доменные ошибки.
package domain

import (
	"errors"
	"time"
)

var (
	// ErrUserIDRequired is returned when user_id is missing or zero.
	// ErrUserIDRequired возвращается, когда user_id отсутствует или равен нулю.
	ErrUserIDRequired = errors.New("user_id is required")
	// ErrOrderIDRequired is returned when order_id is missing or zero.
	// ErrOrderIDRequired возвращается, когда order_id отсутствует или равен нулю.
	ErrOrderIDRequired = errors.New("order_id is required")
	// ErrOrderNotFound is returned when the order does not exist.
	// ErrOrderNotFound возвращается, когда заказ не найден.
	ErrOrderNotFound = errors.New("order not found")
	// ErrCartEmpty is returned when the cart has no items.
	// ErrCartEmpty возвращается, когда корзина пуста.
	ErrCartEmpty = errors.New("cart is empty")
	// ErrInvalidPage is returned when page number is less than 1.
	// ErrInvalidPage возвращается, когда номер страницы меньше 1.
	ErrInvalidPage = errors.New("invalid page")
	// ErrInvalidPageSize is returned when page size is less than 1.
	// ErrInvalidPageSize возвращается, когда размер страницы меньше 1.
	ErrInvalidPageSize = errors.New("invalid page size")
	// ErrOrderNotPending is returned when the order is not in pending status.
	// ErrOrderNotPending возвращается, когда заказ не в статусе pending.
	ErrOrderNotPending = errors.New("order is not pending")
)

// OrderStatus represents the lifecycle state of an order.
// OrderStatus представляет состояние жизненного цикла заказа.
type OrderStatus string

const (
	// OrderStatusPending means the order is created but not yet paid.
	// OrderStatusPending означает, что заказ создан, но ещё не оплачен.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid means the order has been successfully paid.
	// OrderStatusPaid означает, что заказ успешно оплачен.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusFailed means payment or fulfillment failed.
	// OrderStatusFailed означает, что оплата или выполнение не удались.
	OrderStatusFailed OrderStatus = "failed"
	// OrderStatusCancelled means the order was cancelled by the user.
	// OrderStatusCancelled означает, что заказ отменён пользователем.
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem is a single line item within an order.
// OrderItem — одна позиция в составе заказа.
type OrderItem struct {
	// ProductID is the catalog product identifier for this line.
	// ProductID — идентификатор товара в каталоге для этой позиции.
	ProductID int64
	// Quantity is the number of units ordered.
	// Quantity — количество заказанных единиц.
	Quantity int32
	// Name is the product display name snapshot at order time.
	// Name — снимок названия товара на момент создания заказа.
	Name string
	// Price is the unit price snapshot at order time in minor currency units.
	// Price — снимок цены за единицу на момент заказа в минимальных единицах валюты.
	Price int64
}

// Order is the aggregate root for a customer purchase.
// Order — корневая сущность покупки клиента.
type Order struct {
	// ID is the auto-generated primary key of the order.
	// ID — автогенерируемый первичный ключ заказа.
	ID int64
	// UserID is the customer who placed the order.
	// UserID — клиент, оформивший заказ.
	UserID int64
	// Items is the list of product lines in this order.
	// Items — список товарных позиций в заказе.
	Items []OrderItem
	// TotalPrice is the sum of (Price * Quantity) across all items.
	// TotalPrice — сумма (Price * Quantity) по всем позициям.
	TotalPrice int64
	// Status reflects the current order lifecycle state.
	// Status отражает текущее состояние жизненного цикла заказа.
	Status OrderStatus
	// PaymentID links to the payment record once the order is paid (0 if unpaid).
	// PaymentID ссылается на запись платежа после оплаты (0 если не оплачен).
	PaymentID int64
	// CreatedAt is the timestamp when the order was inserted.
	// CreatedAt — метка времени создания заказа.
	CreatedAt time.Time
}

// RecalcTotal recomputes TotalPrice from item prices and quantities.
// RecalcTotal пересчитывает TotalPrice из цен и количества позиций.
func (o *Order) RecalcTotal() {
	var total int64
	for _, item := range o.Items {
		total += item.Price * int64(item.Quantity)
	}
	o.TotalPrice = total
}
