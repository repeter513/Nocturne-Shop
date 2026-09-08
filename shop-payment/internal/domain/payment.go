package domain

import (
	"errors"
	"time"
)

var (
	ErrPaymentIDRequired = errors.New("payment_id is required")
	ErrOrderIDRequired   = errors.New("order_id is required")
	ErrUserIDRequired    = errors.New("user_id is required")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrInvalidPage       = errors.New("invalid page")
	ErrInvalidPageSize   = errors.New("invalid page size")
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type Payment struct {
	ID        int64
	OrderID   int64
	UserID    int64
	Amount    float64
	Status    PaymentStatus
	CreatedAt time.Time
}
