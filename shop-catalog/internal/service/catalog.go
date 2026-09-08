package service

import (
	"context"
	"fmt"

	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
)

type CatalogService struct {
	productRepo     repository.ProductRepository
	reservationRepo repository.ReservationRepository
	stockManager    repository.AtomicStockManager
}

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

func (s *CatalogService) ListProducts(ctx context.Context, params repository.ListProductsParams) ([]*domain.Product, int32, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}
	return s.productRepo.List(ctx, params)
}

func (s *CatalogService) ListCategories(ctx context.Context, parentID *int64) ([]*domain.Category, error) {
	return s.productRepo.ListCategories(ctx, parentID)
}

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

func (s *CatalogService) GetAvailableStock(ctx context.Context, productID int64) (int32, error) {
	stock, err := s.GetStock(ctx, []int64{productID})
	if err != nil {
		return 0, err
	}
	return stock[productID], nil
}

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

func (s *CatalogService) CleanupExpiredReservations(ctx context.Context) error {
	return s.reservationRepo.ExpireOverdue(ctx)
}

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
