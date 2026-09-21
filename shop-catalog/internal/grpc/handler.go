// Package grpc implements the CatalogService gRPC API.
// Пакет grpc реализует gRPC API CatalogService.
package grpc

import (
	"context"
	"errors"
	"strconv"

	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
	"github.com/repeter513/shop-catalog/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler serves catalog RPC methods backed by CatalogService.
// Handler обслуживает RPC-методы каталога через CatalogService.
type Handler struct {
	catalogv1.UnimplementedCatalogServiceServer
	// svc contains product listing, stock, and reservation business logic.
	// svc содержит бизнес-логику каталога, остатков и резервов.
	svc *service.CatalogService
	// defaultTTL is reservation TTL in seconds passed to ReserveStock when client omits TTL.
	// defaultTTL — TTL резерва в секундах для ReserveStock когда клиент не передаёт TTL.
	defaultTTL int32
}

// NewHandler creates a gRPC handler with the given service and default TTL.
// NewHandler создаёт gRPC handler с сервисом и TTL по умолчанию.
func NewHandler(svc *service.CatalogService, defaultTTL int32) *Handler {
	return &Handler{svc: svc, defaultTTL: defaultTTL}
}

// GetProduct returns a single product by ID.
// GetProduct возвращает один товар по ID.
func (h *Handler) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	product, err := h.svc.GetProduct(ctx, req.GetProductId())
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.GetProductResponse{Product: toProtoProduct(product)}, nil
}

// ListProducts returns a paginated product list.
// ListProducts возвращает постраничный список товаров.
// Zero page/pageSize from client are normalized in service (defaults page=1, pageSize=20, max 100).
// Нулевые page/pageSize от клиента нормализуются в service (дефолты page=1, pageSize=20, макс. 100).
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

// ListCategories returns all product categories.
// ListCategories возвращает все категории товаров.
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

// GetStock returns available stock for a product.
// GetStock возвращает доступный остаток товара.
func (h *Handler) GetStock(ctx context.Context, req *catalogv1.GetStockRequest) (*catalogv1.GetStockResponse, error) {
	qty, err := h.svc.GetAvailableStock(ctx, req.GetProductId())
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.GetStockResponse{
		Stock: &catalogv1.Stock{ProductId: req.GetProductId(), Quantity: qty},
	}, nil
}

// ReserveStock reserves quantity for an order.
// ReserveStock резервирует количество для заказа.
// Uses handler defaultTTL from config (RESERVATION_TTL env, default 5m).
// Использует defaultTTL handler из config (env RESERVATION_TTL, дефолт 5m).
func (h *Handler) ReserveStock(ctx context.Context, req *catalogv1.ReserveStockRequest) (*catalogv1.ReserveStockResponse, error) {
	res, err := h.svc.ReserveStock(ctx, map[int64]int32{
		req.GetProductId(): req.GetQuantity(),
	}, req.GetOrderId(), h.defaultTTL)
	if err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ReserveStockResponse{Reservation: toProtoReservation(res, req.GetProductId(), req.GetQuantity())}, nil
}

// ReleaseStock cancels a stock reservation.
// ReleaseStock снимает резерв остатков.
// Invalid reservation_id string → InvalidArgument before service layer.
// Невалидная строка reservation_id → InvalidArgument до service слоя.
func (h *Handler) ReleaseStock(ctx context.Context, req *catalogv1.ReleaseStockRequest) (*catalogv1.ReleaseStockResponse, error) {
	id, err := strconv.ParseInt(req.GetReservationId(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reservation_id")
	}
	if err := h.svc.ReleaseStock(ctx, id, nil, 0); err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ReleaseStockResponse{Reservation: &catalogv1.Reservation{ReservationId: req.GetReservationId()}}, nil
}

// ConfirmReservation confirms a reservation and deducts physical stock.
// ConfirmReservation подтверждает резерв и списывает физический остаток.
func (h *Handler) ConfirmReservation(ctx context.Context, req *catalogv1.ConfirmReservationRequest) (*catalogv1.ConfirmReservationResponse, error) {
	id, err := strconv.ParseInt(req.GetReservationId(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reservation_id")
	}
	if err := h.svc.ConfirmReservation(ctx, id, 0); err != nil {
		return nil, mapError(err)
	}
	return &catalogv1.ConfirmReservationResponse{Reservation: &catalogv1.Reservation{ReservationId: req.GetReservationId()}}, nil
}

// toProtoReservation maps a domain reservation to protobuf.
// toProtoReservation преобразует доменный резерв в protobuf.
func toProtoReservation(res *domain.StockReservation, productID int64, qty int32) *catalogv1.Reservation {
	if res == nil {
		return nil
	}
	return &catalogv1.Reservation{
		ReservationId: strconv.FormatInt(res.ID, 10),
		OrderId:       res.OrderID,
		ProductId:     productID,
		Quantity:      qty,
		Status:        catalogv1.ReservationStatus_RESERVATION_STATUS_PENDING,
	}
}

// toProtoProduct maps a domain product to protobuf.
// toProtoProduct преобразует доменный товар в protobuf.
func toProtoProduct(p *domain.Product) *catalogv1.Product {
	if p == nil {
		return nil
	}
	return &catalogv1.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       int64(p.Price),
		CategoryId:  p.CategoryID,
		Active:      p.Active,
	}
}

// toProtoCategory maps a domain category to protobuf.
// toProtoCategory преобразует доменную категорию в protobuf.
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

// mapError converts domain errors to gRPC status codes.
// mapError преобразует доменные ошибки в gRPC status codes.
// NotFound: missing product/category/reservation. FailedPrecondition: stock/reservation state. InvalidArgument: bad input.
// NotFound: отсутствие product/category/reservation. FailedPrecondition: состояние stock/reservation. InvalidArgument: неверный ввод.
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
