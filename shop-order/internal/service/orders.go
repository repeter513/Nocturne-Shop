package service

import (
	"context"

	"github.com/repeter513/shop-order/internal/adapter/cart"
	"github.com/repeter513/shop-order/internal/adapter/catalog"
	"github.com/repeter513/shop-order/internal/adapter/payment"
	"github.com/repeter513/shop-order/internal/domain"
	"github.com/repeter513/shop-order/internal/repository"
)

type OrderService struct {
	repo    repository.OrderRepository
	cart    *cart.Client
	catalog *catalog.Client
	payment *payment.Client
}

func NewOrderService(repo repository.OrderRepository, cart *cart.Client, catalog *catalog.Client, payment *payment.Client) *OrderService {
	return &OrderService{repo: repo, cart: cart, catalog: catalog, payment: payment}
}

func (s *OrderService) GetOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}
	if orderID == 0 {
		return nil, domain.ErrOrderIDRequired
	}
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserID != userID {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) ListOrders(
	ctx context.Context,
	userID int64,
	page, pageSize int32,
) ([]domain.Order, int32, error) {
	if userID == 0 {
		return nil, 0, domain.ErrUserIDRequired
	}
	if page < 1 {
		return nil, 0, domain.ErrInvalidPage
	}
	if pageSize < 1 {
		return nil, 0, domain.ErrInvalidPageSize
	}
	return s.repo.ListByUser(ctx, userID, page, pageSize)
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64) (*domain.Order, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}

	cartData, err := s.cart.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cartData == nil || len(cartData.Items) == 0 {
		return nil, domain.ErrCartEmpty
	}

	order := &domain.Order{
		UserID: userID,
		Status: domain.OrderStatusPending,
		Items:  cartItemsToOrderItems(cartData.Items),
	}
	order.RecalcTotal()

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	for _, item := range cartData.Items {
		if err := s.catalog.ReserveStock(ctx, order.ID, item.ProductID, item.Quantity); err != nil {
			s.failOrder(ctx, order.ID, cartData.Items, 0)
			return nil, err
		}
	}

	return order, nil
}

func (s *OrderService) PayOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := s.ownedPendingOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}

	paymentID, err := s.payment.CreatePayment(ctx, order.ID, userID, order.TotalPrice)
	if err != nil {
		s.failOrder(ctx, order.ID, orderItemsToCartItems(order.Items), paymentID)
		return nil, err
	}

	if err := s.catalog.ConfirmReservation(ctx, order.ID); err != nil {
		s.failOrder(ctx, order.ID, orderItemsToCartItems(order.Items), paymentID)
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid, paymentID); err != nil {
		return nil, err
	}

	_ = s.cart.ClearCart(ctx, userID)

	order.Status = domain.OrderStatusPaid
	order.PaymentID = paymentID
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := s.ownedPendingOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}

	s.releaseReserved(ctx, order.ID, orderItemsToCartItems(order.Items))
	if err := s.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled, 0); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	order.PaymentID = 0
	return order, nil
}

func (s *OrderService) ownedPendingOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	if orderID == 0 {
		return nil, domain.ErrOrderIDRequired
	}
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, domain.ErrOrderNotFound
	}
	if order.UserID != userID {
		return nil, domain.ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPending {
		return nil, domain.ErrOrderNotPending
	}
	return order, nil
}

func cartItemsToOrderItems(items []cart.Item) []domain.OrderItem {
	out := make([]domain.OrderItem, 0, len(items))
	for _, item := range items {
		out = append(out, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Name:      item.Name,
			Price:     item.Price,
		})
	}
	return out
}

func orderItemsToCartItems(items []domain.OrderItem) []cart.Item {
	out := make([]cart.Item, 0, len(items))
	for _, item := range items {
		out = append(out, cart.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Name:      item.Name,
			Price:     item.Price,
		})
	}
	return out
}

func (s *OrderService) failOrder(ctx context.Context, orderID int64, reserved []cart.Item, paymentID int64) {
	s.releaseReserved(ctx, orderID, reserved)
	_ = s.repo.UpdateStatus(ctx, orderID, domain.OrderStatusFailed, paymentID)
}

func (s *OrderService) releaseReserved(ctx context.Context, orderID int64, items []cart.Item) {
	for _, item := range items {
		_ = s.catalog.ReleaseStock(ctx, orderID, item.ProductID, item.Quantity)
	}
}
