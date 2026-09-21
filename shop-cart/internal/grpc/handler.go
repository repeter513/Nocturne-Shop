// Package grpc implements the cart gRPC API handlers.
// Пакет grpc реализует gRPC-обработчики API корзины.
package grpc

import (
	"context"
	"errors"

	"github.com/repeter513/shop-cart/internal/domain"
	"github.com/repeter513/shop-cart/internal/service"
	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler serves cart RPC methods.
// Handler обслуживает RPC-методы корзины.
type Handler struct {
	cartv1.UnimplementedCartServiceServer
	// svc contains cart business logic invoked by each RPC method.
	// svc содержит бизнес-логику корзины, вызываемую каждым RPC-методом.
	svc *service.CartService
}

// NewHandler creates a gRPC handler backed by the cart service.
// NewHandler создаёт gRPC-обработчик на основе сервиса корзины.
func NewHandler(svc *service.CartService) *Handler {
	return &Handler{svc: svc}
}

// GetCart returns the authenticated user's cart.
// GetCart возвращает корзину аутентифицированного пользователя.
func (h *Handler) GetCart(ctx context.Context, _ *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	// Auth: extract user_id from JWT placed in context by UnaryServerInterceptor.
	// Аутентификация: извлечение user_id из JWT, помещённого в контекст UnaryServerInterceptor.
	userID, err := userID(ctx)
	if err != nil {
		return nil, err // codes.Unauthenticated — missing or invalid JWT
	}
	cart, err := h.svc.GetCart(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &cartv1.GetCartResponse{Cart: toProtoCart(cart)}, nil
}

// AddToCart adds a product to the authenticated user's cart.
// AddToCart добавляет товар в корзину аутентифицированного пользователя.
func (h *Handler) AddToCart(ctx context.Context, req *cartv1.AddToCartRequest) (*cartv1.AddToCartResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := h.svc.AddToCart(ctx, userID, int(req.GetProductId()), req.GetQuantity())
	if err != nil {
		return nil, mapError(err)
	}
	return &cartv1.AddToCartResponse{Cart: toProtoCart(cart)}, nil
}

// UpdateCartItem updates the quantity of a cart item.
// UpdateCartItem обновляет количество позиции в корзине.
func (h *Handler) UpdateCartItem(ctx context.Context, req *cartv1.UpdateCartItemRequest) (*cartv1.UpdateCartItemResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := h.svc.UpdateCartItem(ctx, userID, int(req.GetProductId()), req.GetQuantity())
	if err != nil {
		return nil, mapError(err)
	}
	return &cartv1.UpdateCartItemResponse{Cart: toProtoCart(cart)}, nil
}

// RemoveFromCart removes a product from the authenticated user's cart.
// RemoveFromCart удаляет товар из корзины аутентифицированного пользователя.
func (h *Handler) RemoveFromCart(ctx context.Context, req *cartv1.RemoveFromCartRequest) (*cartv1.RemoveFromCartResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := h.svc.RemoveFromCart(ctx, userID, int(req.GetProductId()))
	if err != nil {
		return nil, mapError(err)
	}
	return &cartv1.RemoveFromCartResponse{Cart: toProtoCart(cart)}, nil
}

// ClearCart removes all items from the authenticated user's cart.
// ClearCart удаляет все позиции из корзины аутентифицированного пользователя.
func (h *Handler) ClearCart(ctx context.Context, _ *cartv1.ClearCartRequest) (*cartv1.ClearCartResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ClearCart(ctx, userID); err != nil {
		return nil, mapError(err)
	}
	return &cartv1.ClearCartResponse{}, nil
}

// userID extracts the authenticated user ID from the request context.
// userID извлекает ID аутентифицированного пользователя из контекста запроса.
//
// The JWT interceptor (pkgauth.UnaryServerInterceptor) validates the token
// and stores the user_id in context before this handler runs.
// JWT-интерцептор (pkgauth.UnaryServerInterceptor) проверяет токен
// и сохраняет user_id в контексте до вызова обработчика.
//
// Returns codes.Unauthenticated when the token is missing, expired, or invalid.
// Возвращает codes.Unauthenticated, когда токен отсутствует, просрочен или невалиден.
func userID(ctx context.Context) (int, error) {
	id, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return int(id), nil
}

// toProtoCart converts a domain cart to its protobuf representation.
// toProtoCart преобразует доменную корзину в protobuf-представление.
func toProtoCart(c *domain.Cart) *cartv1.Cart {
	if c == nil {
		return nil
	}
	items := make([]*cartv1.CartItem, 0, len(c.Items))
	for i := range c.Items {
		items = append(items, toProtoCartItem(&c.Items[i]))
	}
	return &cartv1.Cart{
		UserId:     int64(c.UserId),
		Items:      items,
		TotalPrice: c.Total,
	}
}

// toProtoCartItem converts a domain cart item to protobuf.
// toProtoCartItem преобразует доменную позицию корзины в protobuf.
func toProtoCartItem(item *domain.CartItems) *cartv1.CartItem {
	if item == nil {
		return nil
	}
	return &cartv1.CartItem{
		ProductId: int64(item.ProductId),
		Quantity:  int32(item.Quantity),
		Name:      item.Name,
		Price:     item.Price,
	}
}

// mapError translates domain errors into gRPC status codes.
// mapError преобразует доменные ошибки в коды статуса gRPC.
//
// Status code mapping:
// Сопоставление кодов статуса:
//   - InvalidArgument  — bad input (missing IDs, invalid quantity)
//   - NotFound         — cart item or product does not exist
//   - FailedPrecondition — insufficient stock for requested quantity
//   - Internal         — unexpected database or downstream errors
//
//   - InvalidArgument  — неверный ввод (отсутствующие ID, невалидное количество)
//   - NotFound         — позиция корзины или товар не найдены
//   - FailedPrecondition — недостаточный остаток для запрошенного количества
//   - Internal         — неожиданные ошибки БД или downstream-сервисов
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrIDRequired),
		errors.Is(err, domain.ErrProductRequired),
		errors.Is(err, domain.ErrInvalidQuantity):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrCartItemNotFound),
		errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
