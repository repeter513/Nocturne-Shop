// Package domain defines cart entities and domain errors.
// Пакет domain определяет сущности корзины и доменные ошибки.
package domain

import "errors"

// Domain errors for cart operations.
// Доменные ошибки операций с корзиной.
var (
	// ErrIDRequired is returned when user_id is zero or missing.
	// ErrIDRequired возвращается, когда user_id равен нулю или отсутствует.
	ErrIDRequired = errors.New("id is required")
	// ErrProductRequired is returned when product_id is zero or missing.
	// ErrProductRequired возвращается, когда product_id равен нулю или отсутствует.
	ErrProductRequired = errors.New("product is required")
	// ErrInvalidQuantity is returned when quantity is zero or negative.
	// ErrInvalidQuantity возвращается, когда количество равно нулю или отрицательно.
	ErrInvalidQuantity = errors.New("invalid quantity")
	// ErrCartItemNotFound is returned when the cart line does not exist.
	// ErrCartItemNotFound возвращается, когда позиция корзины не найдена.
	ErrCartItemNotFound = errors.New("cart item not found")
	// ErrProductNotFound is returned when the product is missing or inactive in catalog.
	// ErrProductNotFound возвращается, когда товар отсутствует или неактивен в каталоге.
	ErrProductNotFound = errors.New("product not found")
	// ErrInsufficientStock is returned when requested quantity exceeds available stock.
	// ErrInsufficientStock возвращается, когда запрошенное количество превышает остаток.
	ErrInsufficientStock = errors.New("insufficient stock")
)

// CartItems represents a single line item in a user's cart.
// CartItems представляет одну позицию в корзине пользователя.
type CartItems struct {
	// ProductId is the catalog product identifier for this line.
	// ProductId — идентификатор товара в каталоге для этой позиции.
	ProductId int
	// Quantity is the number of units the user wants to purchase.
	// Quantity — количество единиц товара, которое хочет купить пользователь.
	Quantity int
	// Name is the product display name enriched from the catalog service.
	// Name — отображаемое название товара, обогащённое из сервиса каталога.
	Name string
	// Price is the unit price in minor currency units (e.g. cents).
	// Price — цена за единицу в минимальных единицах валюты (например, копейки).
	Price int64
}

// Cart aggregates all items and total price for a user.
// Cart объединяет все позиции и итоговую сумму для пользователя.
type Cart struct {
	// UserId is the owner of this shopping cart.
	// UserId — владелец данной корзины покупок.
	UserId int
	// Items is the list of product lines currently in the cart.
	// Items — список товарных позиций, находящихся в корзине.
	Items []CartItems
	// Total is the sum of (Price * Quantity) across all items.
	// Total — сумма (Price * Quantity) по всем позициям.
	Total int64
}
