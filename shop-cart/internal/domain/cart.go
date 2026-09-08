package domain

import "errors"

var (
	ErrIDRequired        = errors.New("id is required")
	ErrProductRequired   = errors.New("product is required")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type CartItems struct {
	ProductId int
	Quantity  int
	Name      string
	Price     float64
}
type Cart struct {
	UserId int
	Items  []CartItems
	Total  float64
}
