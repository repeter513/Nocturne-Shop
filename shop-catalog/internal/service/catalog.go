// Package service implements catalog business logic.
// Пакет service реализует бизнес-логику каталога.
package service

import (
	"context"
	"fmt"

	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
)

// CatalogService manages product listings, stock, and reservations.
// CatalogService управляет каталогом товаров, остатками и резервами.
type CatalogService struct {
	// productRepo reads products, categories, and physical stock.
	// productRepo читает товары, категории и физический остаток.
	productRepo repository.ProductRepository
	// reservationRepo queries reservation rows and runs expiry cleanup.
	// reservationRepo запрашивает резервы и выполняет очистку просроченных.
	reservationRepo repository.ReservationRepository
	// stockManager atomically reserves, releases, and confirms stock in transactions.
	// stockManager атомарно резервирует, снимает и подтверждает остаток в транзакциях.
	stockManager repository.AtomicStockManager
}

// NewCatalogService constructs a CatalogService with its dependencies.
// NewCatalogService создаёт CatalogService с зависимостями.
func NewCatalogService(
	productRepo repository.ProductRepository,
	reservationRepo repository.ReservationRepository,
	stockManager repository.AtomicStockManager,
) *CatalogService {
	return &CatalogService{
		productRepo:     productRepo,
		reservationRepo: reservationRepo,
		stockManager:    stockManager,
	}
}

// GetProduct returns a single product by ID.
// GetProduct возвращает один товар по ID.
func (s *CatalogService) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}
	return product, nil
}

// ListProducts returns a paginated product list with total count.
// ListProducts возвращает постраничный список товаров с общим количеством.
// Pagination defaults: page=1 if <=0, pageSize=20 if <=0 or >100 (max 100).
// Дефолты пагинации: page=1 если <=0, pageSize=20 если <=0 или >100 (макс. 100).
func (s *CatalogService) ListProducts(ctx context.Context, params repository.ListProductsParams) ([]*domain.Product, int32, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}
	return s.productRepo.List(ctx, params)
}

// ListCategories returns categories, optionally filtered by parent.
// ListCategories возвращает категории, опционально отфильтрованные по родителю.
func (s *CatalogService) ListCategories(ctx context.Context, parentID *int64) ([]*domain.Category, error) {
	return s.productRepo.ListCategories(ctx, parentID)
}

// GetStock returns physical stock minus active reservations per product.
// GetStock возвращает физический остаток минус активные резервы по каждому товару.
// Available = physical stock - sum(active reservation quantities for product).
// Доступно = физический остаток - сумма(active reservation quantities для товара).
func (s *CatalogService) GetStock(ctx context.Context, productIDs []int64) (map[int64]int32, error) {
	if len(productIDs) == 0 {
		return map[int64]int32{}, nil
	}

	stock, err := s.productRepo.GetStock(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	for _, productID := range productIDs {
		reservations, err := s.reservationRepo.GetActiveReservations(ctx, productID)
		if err != nil {
			return nil, err
		}

		var reserved int32
		for _, res := range reservations {
			if qty, ok := res.Items[productID]; ok {
				reserved += qty
			}
		}

		if current, ok := stock[productID]; ok {
			stock[productID] = current - reserved
		}
	}

	return stock, nil
}

// GetAvailableStock returns available quantity for a single product.
// GetAvailableStock возвращает доступное количество для одного товара.
func (s *CatalogService) GetAvailableStock(ctx context.Context, productID int64) (int32, error) {
	stock, err := s.GetStock(ctx, []int64{productID})
	if err != nil {
		return 0, err
	}
	return stock[productID], nil
}

// ReserveStock atomically reserves items for an order with a TTL.
// ReserveStock атомарно резервирует позиции для заказа с TTL.
// ttlSeconds default 300 (5 min) when <= 0; idempotent merge by order_id in stock manager.
// ttlSeconds по умолчанию 300 (5 мин) если <= 0; идемпотентное объединение по order_id в stock manager.
func (s *CatalogService) ReserveStock(
	ctx context.Context,
	items map[int64]int32,
	orderID int64,
	ttlSeconds int32,
) (*domain.StockReservation, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to reserve")
	}
	if orderID == 0 {
		return nil, domain.ErrOrderIDRequired
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	return s.stockManager.ReserveWithTransaction(ctx, items, orderID, ttlSeconds)
}

// ReleaseStock cancels a reservation by ID or by order items.
// ReleaseStock снимает резерв по ID или по позициям заказа.
func (s *CatalogService) ReleaseStock(
	ctx context.Context,
	reservationID int64,
	items map[int64]int32,
	orderID int64,
) error {
	if reservationID == 0 && len(items) == 0 {
		return fmt.Errorf("either reservation_id or items must be provided")
	}
	if reservationID != 0 {
		return s.releaseByReservationID(ctx, reservationID)
	}
	return s.releaseByItems(ctx, items, orderID)
}

// ConfirmReservation permanently deducts stock for a confirmed order.
// ConfirmReservation окончательно списывает остаток по подтверждённому заказу.
// Error paths: not found, expired, already released → domain errors mapped in gRPC layer.
// Пути ошибок: not found, expired, already released → доменные ошибки мапятся в gRPC слое.
func (s *CatalogService) ConfirmReservation(ctx context.Context, reservationID, orderID int64) error {
	reservation, err := s.findReservation(ctx, reservationID, orderID)
	if err != nil {
		return err
	}
	if err := validateReservationActive(reservation); err != nil {
		return err
	}
	return s.stockManager.ConfirmWithTransaction(ctx, reservation.ID)
}

// CleanupExpiredReservations marks overdue active reservations as expired.
// CleanupExpiredReservations помечает просроченные активные резервы как expired.
func (s *CatalogService) CleanupExpiredReservations(ctx context.Context) error {
	return s.reservationRepo.ExpireOverdue(ctx)
}

// releaseByReservationID releases all or specified items by reservation ID.
// releaseByReservationID снимает резерв по ID резервации.
func (s *CatalogService) releaseByReservationID(ctx context.Context, reservationID int64) error {
	reservation, err := s.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return err
	}
	if err := validateReservationActive(reservation); err != nil {
		return err
	}
	return s.stockManager.ReleaseWithTransaction(ctx, reservationID, reservation.Items)
}

// releaseByItems releases specific items from the order's active reservation.
// releaseByItems снимает указанные позиции из активного резерва заказа.
func (s *CatalogService) releaseByItems(ctx context.Context, items map[int64]int32, orderID int64) error {
	if orderID == 0 {
		return domain.ErrOrderIDRequired
	}

	reservation, err := s.reservationRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := validateReservationActive(reservation); err != nil {
		return err
	}

	for productID, qty := range items {
		if reserved, ok := reservation.Items[productID]; !ok || reserved < qty {
			return fmt.Errorf("%w: product %d", domain.ErrInsufficientStock, productID)
		}
	}

	return s.stockManager.ReleaseWithTransaction(ctx, reservation.ID, items)
}

// findReservation looks up a reservation by ID or order ID.
// findReservation ищет резерв по ID резервации или ID заказа.
func (s *CatalogService) findReservation(ctx context.Context, reservationID, orderID int64) (*domain.StockReservation, error) {
	switch {
	case reservationID != 0:
		reservation, err := s.reservationRepo.FindByID(ctx, reservationID)
		if err != nil {
			return nil, err
		}
		if reservation == nil {
			return nil, domain.ErrReservationNotFound
		}
		return reservation, nil
	case orderID != 0:
		reservation, err := s.reservationRepo.FindByOrderID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		if reservation == nil {
			return nil, domain.ErrReservationNotFound
		}
		return reservation, nil
	default:
		return nil, fmt.Errorf("either reservation_id or order_id must be provided")
	}
}

// validateReservationActive checks that the reservation can still be modified.
// validateReservationActive проверяет, что резерв ещё можно изменить.
func validateReservationActive(reservation *domain.StockReservation) error {
	if reservation == nil {
		return domain.ErrReservationNotFound
	}
	switch reservation.Status {
	case domain.ReservationReleased, domain.ReservationConfirmed:
		return domain.ErrReservationAlreadyReleased
	case domain.ReservationExpired:
		return domain.ErrReservationExpired
	default:
		return nil
	}
}
