package domain

import (
	"errors"
	"time"
)

var (
	ErrProductNotFound            = errors.New("product not found")
	ErrInsufficientStock          = errors.New("insufficient stock")
	ErrCategoryNotFound           = errors.New("category not found")
	ErrReservationNotFound        = errors.New("reservation not found")
	ErrReservationExpired         = errors.New("reservation expired")
	ErrReservationAlreadyReleased = errors.New("reservation already released")
	ErrOrderIDRequired            = errors.New("order_id is required")
)

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	CategoryID  int64
	Stock       int32
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID          int64
	Name        string
	Description string
	ParentID    *int64
	CreatedAt   time.Time
}

type StockReservation struct {
	ID        int64
	OrderID   int64
	Items     map[int64]int32
	TTL       time.Duration
	ExpiresAt time.Time
	CreatedAt time.Time
	Status    ReservationStatus
}

type ReservationStatus string

const (
	ReservationActive            ReservationStatus = "active"
	ReservationReleased          ReservationStatus = "released"
	ReservationConfirmed         ReservationStatus = "confirmed"
	ReservationExpired           ReservationStatus = "expired"
	ReservationPartiallyReleased ReservationStatus = "partially_released"
)
