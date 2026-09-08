package domain
import (
	"errors"
	"time"
)
var (
	ErrUserIDRequired  = errors.New("user_id is required")
	ErrOrderIDRequired = errors.New("order_id is required")
	ErrOrderNotFound   = errors.New("order not found")
	ErrCartEmpty       = errors.New("cart is empty")
	ErrInvalidPage     = errors.New("invalid page")
	ErrInvalidPageSize = errors.New("invalid page size")
	ErrOrderNotPending = errors.New("order is not pending")
)
type OrderStatus string
const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)
type OrderItem struct {
	ProductID int64
	Quantity  int32
	Name      string
	Price     float64
}
type Order struct {
	ID         int64
	UserID     int64
	Items      []OrderItem
	TotalPrice float64
	Status     OrderStatus
	PaymentID  int64
	CreatedAt  time.Time
}
func (o *Order) RecalcTotal() {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalPrice = total
}
