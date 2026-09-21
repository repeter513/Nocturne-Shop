// Package domain defines payment entities and domain errors.
// Пакет domain определяет сущности платежей и доменные ошибки.
package domain

import (
	"errors"
	"time"
)

// Domain errors for payment operations.
// Доменные ошибки операций с платежами.
var (
	// ErrPaymentIDRequired is returned when payment_id is zero or missing.
	// ErrPaymentIDRequired возвращается, когда payment_id равен нулю или отсутствует.
	ErrPaymentIDRequired = errors.New("payment_id is required")
	// ErrOrderIDRequired is returned when order_id is zero or missing.
	// ErrOrderIDRequired возвращается, когда order_id равен нулю или отсутствует.
	ErrOrderIDRequired = errors.New("order_id is required")
	// ErrUserIDRequired is returned when user_id is zero or missing.
	// ErrUserIDRequired возвращается, когда user_id равен нулю или отсутствует.
	ErrUserIDRequired = errors.New("user_id is required")
	// ErrInvalidAmount is returned when the payment amount is zero or negative.
	// ErrInvalidAmount возвращается, когда сумма платежа равна нулю или отрицательна.
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrPaymentNotFound is returned when the payment record does not exist.
	// ErrPaymentNotFound возвращается, когда запись платежа не найдена.
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrInvalidPage is returned when page number is less than 1.
	// ErrInvalidPage возвращается, когда номер страницы меньше 1.
	ErrInvalidPage = errors.New("invalid page")
	// ErrInvalidPageSize is returned when page size is less than 1.
	// ErrInvalidPageSize возвращается, когда размер страницы меньше 1.
	ErrInvalidPageSize = errors.New("invalid page size")
)

// PaymentStatus represents the lifecycle state of a payment.
// PaymentStatus представляет состояние жизненного цикла платежа.
type PaymentStatus string

// Payment status constants.
// Константы статусов платежа.
const (
	// PaymentStatusPending means the payment is initiated but not yet confirmed.
	// PaymentStatusPending означает, что платёж инициирован, но ещё не подтверждён.
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusSuccess means the payment was processed successfully.
	// PaymentStatusSuccess означает, что платёж успешно обработан.
	PaymentStatusSuccess PaymentStatus = "success"
	// PaymentStatusFailed means the payment attempt failed.
	// PaymentStatusFailed означает, что попытка оплаты не удалась.
	PaymentStatusFailed PaymentStatus = "failed"
)

// Payment holds a recorded payment for an order.
// Payment хранит зарегистрированный платёж по заказу.
type Payment struct {
	// ID is the auto-generated primary key of the payment record.
	// ID — автогенерируемый первичный ключ записи платежа.
	ID int64
	// OrderID links this payment to the originating order.
	// OrderID связывает платёж с исходным заказом.
	OrderID int64
	// UserID is the customer who initiated the payment.
	// UserID — клиент, инициировавший платёж.
	UserID int64
	// Amount is the charged sum in major currency units (e.g. dollars).
	// Amount — списанная сумма в основных единицах валюты (например, доллары).
	Amount float64
	// Status reflects the current payment lifecycle state.
	// Status отражает текущее состояние жизненного цикла платежа.
	Status PaymentStatus
	// CreatedAt is the timestamp when the payment record was inserted.
	// CreatedAt — метка времени вставки записи платежа.
	CreatedAt time.Time
}
