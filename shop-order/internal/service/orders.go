// Package service implements order business logic.
// Пакет service реализует бизнес-логику заказов.
package service

import (
	"context"

	"github.com/repeter513/shop-order/internal/adapter/cart"
	"github.com/repeter513/shop-order/internal/adapter/catalog"
	"github.com/repeter513/shop-order/internal/adapter/payment"
	"github.com/repeter513/shop-order/internal/domain"
	"github.com/repeter513/shop-order/internal/repository"
)

// OrderService orchestrates order creation, payment, and cancellation.
// OrderService координирует создание, оплату и отмену заказов.
type OrderService struct {
	// repo persists order aggregates in PostgreSQL.
	// repo сохраняет агрегаты заказов в PostgreSQL.
	repo repository.OrderRepository
	// cart is the downstream gRPC client to read and clear the user's cart.
	// cart — downstream gRPC-клиент для чтения и очистки корзины пользователя.
	cart *cart.Client
	// catalog is the downstream gRPC client for stock reservation and confirmation.
	// catalog — downstream gRPC-клиент для резервирования и подтверждения остатков.
	catalog *catalog.Client
	// payment is the downstream gRPC client to charge the customer.
	// payment — downstream gRPC-клиент для списания оплаты с клиента.
	payment *payment.Client
}

// NewOrderService constructs an OrderService with its dependencies.
// NewOrderService создаёт OrderService с его зависимостями.
func NewOrderService(repo repository.OrderRepository, cart *cart.Client, catalog *catalog.Client, payment *payment.Client) *OrderService {
	return &OrderService{repo: repo, cart: cart, catalog: catalog, payment: payment}
}

// GetOrder returns an order owned by the given user.
// GetOrder возвращает заказ, принадлежащий указанному пользователю.
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

// ListOrders returns paginated orders for a user.
// ListOrders возвращает постраничный список заказов пользователя.
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

// CreateOrder builds an order from the user's cart and reserves stock.
// CreateOrder создаёт заказ из корзины пользователя и резервирует товар.
func (s *OrderService) CreateOrder(ctx context.Context, userID int64) (*domain.Order, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}

	// Step 1: downstream call — cart.GetCart to fetch current cart lines.
	// Шаг 1: downstream-вызов — cart.GetCart для получения текущих позиций корзины.
	cartData, err := s.cart.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cartData == nil || len(cartData.Items) == 0 {
		return nil, domain.ErrCartEmpty
	}

	// Step 2: build pending order aggregate from cart snapshot.
	// Шаг 2: формирование агрегата заказа в статусе pending из снимка корзины.
	order := &domain.Order{
		UserID: userID,
		Status: domain.OrderStatusPending,
		Items:  cartItemsToOrderItems(cartData.Items),
	}
	order.RecalcTotal()

	// Step 3: persist order and line items in a single transaction.
	// Шаг 3: сохранение заказа и позиций в одной транзакции.
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Step 4: downstream call — catalog.ReserveStock for each cart line.
	// Шаг 4: downstream-вызов — catalog.ReserveStock для каждой позиции корзины.
	for _, item := range cartData.Items {
		if err := s.catalog.ReserveStock(ctx, order.ID, item.ProductID, item.Quantity); err != nil {
			// Compensation: mark order failed and release any partial reservations.
			// Компенсация: пометить заказ failed и освободить частичные резервирования.
			s.failOrder(ctx, order.ID, cartData.Items, 0)
			return nil, err
		}
	}

	return order, nil
}

// PayOrder charges the user and confirms stock reservations.
// PayOrder списывает оплату и подтверждает резервирование товара.
func (s *OrderService) PayOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	// Step 1: load and verify the order is pending and owned by the user.
	// Шаг 1: загрузка и проверка, что заказ pending и принадлежит пользователю.
	order, err := s.ownedPendingOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}

	// Step 2: downstream call — payment.CreatePayment to charge the order total.
	// Шаг 2: downstream-вызов — payment.CreatePayment для списания суммы заказа.
	paymentID, err := s.payment.CreatePayment(ctx, order.ID, userID, 0, 0)
	if err != nil {
		s.failOrder(ctx, order.ID, orderItemsToCartItems(order.Items), paymentID)
		return nil, err
	}

	// Step 3: downstream call — catalog.ConfirmReservation to finalize stock deduction.
	// Шаг 3: downstream-вызов — catalog.ConfirmReservation для финального списания остатков.
	if err := s.catalog.ConfirmReservation(ctx, order.ID); err != nil {
		s.failOrder(ctx, order.ID, orderItemsToCartItems(order.Items), paymentID)
		return nil, err
	}

	// Step 4: update order status to paid with payment reference.
	// Шаг 4: обновление статуса заказа на paid со ссылкой на платёж.
	if err := s.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid, paymentID); err != nil {
		return nil, err
	}

	// Step 5: downstream call — cart.ClearCart to empty the user's cart after successful payment.
	// Шаг 5: downstream-вызов — cart.ClearCart для очистки корзины после успешной оплаты.
	_ = s.cart.ClearCart(ctx, userID)

	order.Status = domain.OrderStatusPaid
	order.PaymentID = paymentID
	return order, nil
}

// CancelOrder cancels a pending order and releases reserved stock.
// CancelOrder отменяет заказ в статусе pending и освобождает резерв.
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := s.ownedPendingOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}

	// Step 1: downstream call — catalog.ReleaseOrder to return stock to inventory.
	// Шаг 1: downstream-вызов — catalog.ReleaseOrder для возврата товара на склад.
	s.releaseReserved(ctx, order.ID, orderItemsToCartItems(order.Items))

	// Step 2: update order status to cancelled.
	// Шаг 2: обновление статуса заказа на cancelled.
	if err := s.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled, 0); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	order.PaymentID = 0
	return order, nil
}

// ownedPendingOrder loads a pending order that belongs to the user.
// ownedPendingOrder загружает заказ в статусе pending, принадлежащий пользователю.
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

// cartItemsToOrderItems converts cart items to domain order items.
// cartItemsToOrderItems преобразует позиции корзины в доменные позиции заказа.
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

// orderItemsToCartItems converts domain order items to cart items.
// orderItemsToCartItems преобразует доменные позиции заказа в позиции корзины.
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

// failOrder marks the order as failed and releases reservations.
// failOrder помечает заказ как failed и освобождает резервирования.
func (s *OrderService) failOrder(ctx context.Context, orderID int64, reserved []cart.Item, paymentID int64) {
	s.releaseReserved(ctx, orderID, reserved)
	_ = s.repo.UpdateStatus(ctx, orderID, domain.OrderStatusFailed, paymentID)
}

// releaseReserved releases catalog stock held for an order.
// releaseReserved освобождает товар в каталоге, зарезервированный для заказа.
func (s *OrderService) releaseReserved(ctx context.Context, orderID int64, _ []cart.Item) {
	_ = s.catalog.ReleaseOrder(ctx, orderID)
}
