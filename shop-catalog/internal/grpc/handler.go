package grpc

import (
	"context"
	"errors"

	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
	"github.com/repeter513/shop-catalog/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	catalogv1.UnimplementedCatalogServiceServer
	svc        *service.CatalogService
	defaultTTL int32
}

func NewHandler(svc *service.CatalogService, defaultTTL int32) *Handler {
	return &Handler{svc: svc, defaultTTL: defaultTTL}
}

func (h *Handler) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	product, err := h.svc.GetProduct(ctx, req.GetProductId())
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.GetProductResponse{Product: toProtoProduct(product)}, nil
}

func (h *Handler) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	products, total, err := h.svc.ListProducts(ctx, repository.ListProductsParams{
		Page:       int(req.GetPage()),
		PageSize:   int(req.GetPageSize()),
		CategoryID: req.GetCategoryId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*catalogv1.Product, 0, len(products))
	for _, p := range products {
		out = append(out, toProtoProduct(p))
	}
	return &catalogv1.ListProductsResponse{Products: out, TotalCount: total}, nil
}

func (h *Handler) ListCategories(ctx context.Context, _ *catalogv1.ListCategoriesRequest) (*catalogv1.ListCategoriesResponse, error) {
	categories, err := h.svc.ListCategories(ctx, nil)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*catalogv1.Category, 0, len(categories))
	for _, c := range categories {
		out = append(out, toProtoCategory(c))
	}
	return &catalogv1.ListCategoriesResponse{Categories: out}, nil
}

func (h *Handler) GetStock(ctx context.Context, req *catalogv1.GetStockRequest) (*catalogv1.GetStockResponse, error) {
	qty, err := h.svc.GetAvailableStock(ctx, req.GetProductId())
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.GetStockResponse{
		Stock: &catalogv1.Stock{ProductId: req.GetProductId(), Quantity: qty},
	}, nil
}

func (h *Handler) ReserveStock(ctx context.Context, req *catalogv1.ReserveStockRequest) (*catalogv1.ReserveStockResponse, error) {
	_, err := h.svc.ReserveStock(ctx, map[int64]int32{
		req.GetProductId(): req.GetQuantity(),
	}, req.GetOrderId(), h.defaultTTL)
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ReserveStockResponse{Success: true}, nil
}

func (h *Handler) ReleaseStock(ctx context.Context, req *catalogv1.ReleaseStockRequest) (*catalogv1.ReleaseStockResponse, error) {
	err := h.svc.ReleaseStock(ctx, 0, map[int64]int32{
		req.GetProductId(): req.GetQuantity(),
	}, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ReleaseStockResponse{Success: true}, nil
}

func (h *Handler) ConfirmReservation(ctx context.Context, req *catalogv1.ConfirmReservationRequest) (*catalogv1.ConfirmReservationResponse, error) {
	if err := h.svc.ConfirmReservation(ctx, 0, req.GetOrderId()); err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ConfirmReservationResponse{Success: true}, nil
}

func toProtoProduct(p *domain.Product) *catalogv1.Product {
	if p == nil {
		return nil
	}
	return &catalogv1.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CategoryId:  p.CategoryID,
		Active:      p.Active,
	}
}

func toProtoCategory(c *domain.Category) *catalogv1.Category {
	if c == nil {
		return nil
	}
	out := &catalogv1.Category{
		Id:   c.ID,
		Name: c.Name,
	}
	if c.ParentID != nil {
		out.ParentId = *c.ParentID
	}
	return out
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrReservationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrReservationExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrReservationAlreadyReleased):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrOrderIDRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
