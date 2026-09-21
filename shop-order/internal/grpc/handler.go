// Package grpc exposes order operations over gRPC.
// Пакет grpc предоставляет операции с заказами через gRPC.
package grpc

import (
	"context"
	"errors"

	"github.com/repeter513/shop-order/internal/domain"
	"github.com/repeter513/shop-order/internal/service"
	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the generated OrderServiceServer interface.
// Handler реализует сгенерированный интерфейс OrderServiceServer.
type Handler struct {
	orderv1.UnimplementedOrderServiceServer
	// svc contains order orchestration logic invoked by each RPC method.
	// svc содержит логику оркестрации заказов, вызываемую каждым RPC-методом.
	svc *service.OrderService
}

// NewHandler constructs a gRPC handler backed by OrderService.
// NewHandler создаёт gRPC-обработчик на базе OrderService.
func NewHandler(svc *service.OrderService) *Handler {
	return &Handler{svc: svc}
}

// CreateOrder creates a new order from the authenticated user's cart.
// CreateOrder создаёт новый заказ из корзины аутентифицированного пользователя.
func (h *Handler) CreateOrder(ctx context.Context, _ *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	// Auth: user_id from JWT context identifies the order owner.
	// Аутентификация: user_id из JWT-контекста определяет владельца заказа.
	userID, err := userID(ctx)
	if err != nil {
		return nil, err // codes.Unauthenticated
	}
	order, err := h.svc.CreateOrder(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CreateOrderResponse{Order: toProtoOrder(order)}, nil
}

// PayOrder processes payment for a pending order.
// PayOrder обрабатывает оплату заказа в статусе pending.
func (h *Handler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest) (*orderv1.PayOrderResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := h.svc.PayOrder(ctx, userID, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.PayOrderResponse{Order: toProtoOrder(order)}, nil
}

// CancelOrder cancels a pending order owned by the caller.
// CancelOrder отменяет заказ в статусе pending, принадлежащий вызывающему.
func (h *Handler) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := h.svc.CancelOrder(ctx, userID, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CancelOrderResponse{Order: toProtoOrder(order)}, nil
}

// GetOrder returns a single order by ID for the authenticated user.
// GetOrder возвращает один заказ по ID для аутентифицированного пользователя.
func (h *Handler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := h.svc.GetOrder(ctx, userID, req.GetOrderId())

	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.GetOrderResponse{Order: toProtoOrder(order)}, nil
}

// ListOrders returns paginated orders for the authenticated user.
// ListOrders возвращает постраничный список заказов аутентифицированного пользователя.
func (h *Handler) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page == 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 10
	}

	orders, total, err := h.svc.ListOrders(ctx, userID, page, pageSize)
	if err != nil {
		return nil, mapError(err)
	}

	protoOrders := make([]*orderv1.Order, 0, len(orders))
	for i := range orders {
		protoOrders = append(protoOrders, toProtoOrder(&orders[i]))
	}
	return &orderv1.ListOrdersResponse{
		Orders:     protoOrders,
		TotalCount: total,
	}, nil
}

// userID extracts the authenticated user ID from gRPC context.
// userID извлекает ID аутентифицированного пользователя из gRPC-контекста.
//
// The JWT interceptor validates the token (issuer + AudienceOrder) and stores
// user_id in context before this handler runs.
// JWT-интерцептор проверяет токен (issuer + AudienceOrder) и сохраняет
// user_id в контексте до вызова обработчика.
//
// Returns codes.Unauthenticated when the token is missing, expired, or invalid.
// Возвращает codes.Unauthenticated, когда токен отсутствует, просрочен или невалиден.
func userID(ctx context.Context) (int64, error) {
	id, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return id, nil
}

// toProtoOrder converts a domain order to its protobuf representation.
// toProtoOrder преобразует доменный заказ в protobuf-представление.
func toProtoOrder(o *domain.Order) *orderv1.Order {
	if o == nil {
		return nil
	}
	items := make([]*orderv1.OrderItem, 0, len(o.Items))
	for i := range o.Items {
		items = append(items, toProtoOrderItem(&o.Items[i]))
	}
	return &orderv1.Order{
		OrderId:    o.ID,
		UserId:     o.UserID,
		Items:      items,
		TotalPrice: o.TotalPrice,
		Status:     toProtoStatus(o.Status),
		CreatedAt:  timestamppb.New(o.CreatedAt),
		PaymentId:  o.PaymentID,
	}
}

// toProtoOrderItem converts a domain order item to protobuf.
// toProtoOrderItem преобразует доменную позицию заказа в protobuf.
func toProtoOrderItem(item *domain.OrderItem) *orderv1.OrderItem {
	if item == nil {
		return nil
	}
	return &orderv1.OrderItem{
		ProductId: item.ProductID,
		Quantity:  item.Quantity,
		Name:      item.Name,
		Price:     item.Price,
	}
}

// toProtoStatus maps a domain order status to protobuf enum.
// toProtoStatus сопоставляет доменный статус заказа с protobuf-перечислением.
func toProtoStatus(s domain.OrderStatus) orderv1.OrderStatus {
	switch s {
	case domain.OrderStatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusPaid:
		return orderv1.OrderStatus_ORDER_STATUS_PAID
	case domain.OrderStatusFailed:
		return orderv1.OrderStatus_ORDER_STATUS_FAILED
	case domain.OrderStatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

// mapError translates domain errors to gRPC status codes.
// mapError преобразует доменные ошибки в коды статуса gRPC.
//
// Status code mapping:
// Сопоставление кодов статуса:
//   - InvalidArgument      — bad input (missing IDs, invalid pagination)
//   - NotFound             — order does not exist or belongs to another user
//   - FailedPrecondition   — order not pending, or cart is empty
//   - Internal             — unexpected database or downstream service errors
//
//   - InvalidArgument      — неверный ввод (отсутствующие ID, невалидная пагинация)
//   - NotFound             — заказ не найден или принадлежит другому пользователю
//   - FailedPrecondition   — заказ не pending или корзина пуста
//   - Internal             — неожиданные ошибки БД или downstream-сервисов
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserIDRequired),
		errors.Is(err, domain.ErrOrderIDRequired),
		errors.Is(err, domain.ErrInvalidPage),
		errors.Is(err, domain.ErrInvalidPageSize):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrOrderNotPending):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domain.ErrCartEmpty):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
