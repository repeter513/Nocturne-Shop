// Package domain defines catalog entities and domain errors.
// Пакет domain определяет сущности каталога и доменные ошибки.
package domain

import (
	"errors"
	"time"
)

var (
	// ErrProductNotFound is returned when a product does not exist.
	// ErrProductNotFound возвращается, когда товар не найден.
	ErrProductNotFound = errors.New("product not found")
	// ErrInsufficientStock is returned when available stock is too low.
	// ErrInsufficientStock возвращается при недостаточном остатке.
	ErrInsufficientStock = errors.New("insufficient stock")
	// ErrCategoryNotFound is returned when a category does not exist.
	// ErrCategoryNotFound возвращается, когда категория не найдена.
	ErrCategoryNotFound = errors.New("category not found")
	// ErrReservationNotFound is returned when a reservation does not exist.
	// ErrReservationNotFound возвращается, когда резерв не найден.
	ErrReservationNotFound = errors.New("reservation not found")
	// ErrReservationExpired is returned when a reservation has expired.
	// ErrReservationExpired возвращается, когда резерв просрочен.
	ErrReservationExpired = errors.New("reservation expired")
	// ErrReservationAlreadyReleased is returned when a reservation is no longer active.
	// ErrReservationAlreadyReleased возвращается, когда резерв уже снят или подтверждён.
	ErrReservationAlreadyReleased = errors.New("reservation already released")
	// ErrOrderIDRequired is returned when order_id is missing.
	// ErrOrderIDRequired возвращается, когда order_id не указан.
	ErrOrderIDRequired = errors.New("order_id is required")
)

// Product represents a catalog item with price and stock.
// Product представляет товар каталога с ценой и остатком.
type Product struct {
	// ID is the surrogate primary key (BIGSERIAL).
	// ID — суррогатный первичный ключ (BIGSERIAL).
	ID int64
	// Name is the display title shown in catalog listings.
	// Name — отображаемое название в каталоге.
	Name string
	// Description is long-form product text (may contain newlines).
	// Description — подробное описание товара (может содержать переносы строк).
	Description string
	// Price is unit price in major currency units (stored as NUMERIC in DB).
	// Price — цена за единицу в основных единицах валюты (NUMERIC в БД).
	Price float64
	// CategoryID links to categories.id for filtering and breadcrumbs.
	// CategoryID — ссылка на categories.id для фильтрации и навигации.
	CategoryID int64
	// Stock is physical on-hand quantity; reservations do not decrement until confirm.
	// Stock — физический остаток; резервы не уменьшают его до confirm.
	Stock int32
	// Active=false hides product from ListProducts (soft delete).
	// Active=false скрывает товар из ListProducts (soft delete).
	Active bool
	// CreatedAt is row insertion time (TIMESTAMPTZ).
	// CreatedAt — время вставки строки (TIMESTAMPTZ).
	CreatedAt time.Time
	// UpdatedAt is last modification time (bumped on stock/price changes).
	// UpdatedAt — время последнего изменения (обновляется при изменении остатка/цены).
	UpdatedAt time.Time
}

// Category represents a product category, optionally nested via ParentID.
// Category представляет категорию товаров, опционально вложенную через ParentID.
type Category struct {
	// ID is the category primary key.
	// ID — первичный ключ категории.
	ID int64
	// Name is the category label.
	// Name — название категории.
	Name string
	// Description explains category scope.
	// Description — описание области категории.
	Description string
	// ParentID is nil for root categories; otherwise points to parent row.
	// ParentID nil для корневых категорий; иначе ссылка на родительскую строку.
	ParentID *int64
	// CreatedAt is when the category was added.
	// CreatedAt — когда категория была добавлена.
	CreatedAt time.Time
}

// StockReservation holds reserved quantities for an order until confirm or expiry.
// StockReservation хранит зарезервированные количества для заказа до подтверждения или истечения.
type StockReservation struct {
	// ID is reservation primary key.
	// ID — первичный ключ резерва.
	ID int64
	// OrderID is unique per reservation (one active reservation per order).
	// OrderID уникален на резерв (один активный резерв на заказ).
	OrderID int64
	// Items maps product_id → reserved quantity (JSONB in DB).
	// Items — map product_id → зарезервированное количество (JSONB в БД).
	Items map[int64]int32
	// TTL is the requested hold duration (may differ from ExpiresAt after merge).
	// TTL — запрошенная длительность удержания (может отличаться от ExpiresAt после merge).
	TTL time.Duration
	// ExpiresAt is absolute expiry; cleanup marks status=expired when passed.
	// ExpiresAt — абсолютное время истечения; cleanup помечает status=expired после наступления.
	ExpiresAt time.Time
	// CreatedAt is when the reservation row was first inserted.
	// CreatedAt — когда строка резерва была впервые вставлена.
	CreatedAt time.Time
	// Status tracks lifecycle: active → confirmed|released|expired|partially_released.
	// Status отслеживает жизненный цикл: active → confirmed|released|expired|partially_released.
	Status ReservationStatus
}

// ReservationStatus describes the lifecycle state of a stock reservation.
// ReservationStatus описывает состояние жизненного цикла резерва остатков.
type ReservationStatus string

const (
	// ReservationActive means the reservation is holding stock.
	// ReservationActive означает, что резерв удерживает остаток.
	ReservationActive ReservationStatus = "active"
	// ReservationReleased means all items were released back to available stock.
	// ReservationReleased означает, что все позиции возвращены в доступный остаток.
	ReservationReleased ReservationStatus = "released"
	// ReservationConfirmed means stock was permanently deducted for the order.
	// ReservationConfirmed означает, что остаток списан по заказу.
	ReservationConfirmed ReservationStatus = "confirmed"
	// ReservationExpired means the reservation timed out without confirmation.
	// ReservationExpired означает, что резерв истёк без подтверждения.
	ReservationExpired ReservationStatus = "expired"
	// ReservationPartiallyReleased means some items were released, others remain reserved.
	// ReservationPartiallyReleased означает частичное снятие резерва.
	ReservationPartiallyReleased ReservationStatus = "partially_released"
)
