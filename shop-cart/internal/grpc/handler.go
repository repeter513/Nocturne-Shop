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

type Handler struct {
	cartv1.UnimplementedCartServiceServer
	svc *service.CartService
}

func NewHandler(svc *service.CartService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetCart(ctx context.Context, _ *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := h.svc.GetCart(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &cartv1.GetCartResponse{Cart: toProtoCart(cart)}, nil
}

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

func userID(ctx context.Context) (int, error) {
	id, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return int(id), nil
}

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
