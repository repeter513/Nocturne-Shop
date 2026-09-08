package service

import (
	"context"
	"github.com/repeter513/shop-cart/internal/adapter/catalog"
	"github.com/repeter513/shop-cart/internal/domain"
	"github.com/repeter513/shop-cart/internal/repository"
)

type CartService struct {
	repo    repository.CartItemRepository
	catalog *catalog.Client
}

func NewCartService(
	repo repository.CartItemRepository,
	catalog *catalog.Client) *CartService {
	return &CartService{
		repo:    repo,
		catalog: catalog,
	}
}
func (s *CartService) GetCart(ctx context.Context, userID int) (*domain.Cart, error) {
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	return s.buildCart(ctx, userID)
}

func (s *CartService) AddToCart(ctx context.Context, userID, productID int, quantity int32) (*domain.Cart, error) {
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	if productID == 0 {
		return nil, domain.ErrProductRequired
	}
	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	items, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	newQty := currentQty(items, productID) + int(quantity)
	if err := s.validateProductAndStock(ctx, productID, newQty); err != nil {
		return nil, err
	}

	if err := s.repo.UpsertItems(ctx, userID, productID, int32(newQty)); err != nil {
		return nil, err
	}

	return s.buildCart(ctx, userID)
}

func (s *CartService) UpdateCartItem(ctx context.Context, userID, productID int, quantity int32) (*domain.Cart, error) {
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	if productID == 0 {
		return nil, domain.ErrProductRequired
	}
	if quantity < 0 {
		return nil, domain.ErrInvalidQuantity
	}
	if quantity == 0 {
		return s.RemoveFromCart(ctx, userID, productID)
	}

	if err := s.validateProductAndStock(ctx, productID, int(quantity)); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateItems(ctx, userID, productID, quantity); err != nil {
		return nil, err
	}
	return s.buildCart(ctx, userID)
}

func (s *CartService) RemoveFromCart(ctx context.Context, userID, productID int) (*domain.Cart, error) {
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	if productID == 0 {
		return nil, domain.ErrProductRequired
	}
	if err := s.repo.DeleteItems(ctx, userID, productID); err != nil {
		return nil, err
	}
	return s.buildCart(ctx, userID)
}

func (s *CartService) ClearCart(ctx context.Context, userID int) error {
	if userID == 0 {
		return domain.ErrIDRequired
	}
	return s.repo.Clear(ctx, userID)
}

func (s *CartService) buildCart(ctx context.Context, userID int) (*domain.Cart, error) {
	lines, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	cart := &domain.Cart{
		UserId: userID,
		Items:  make([]domain.CartItems, 0, len(lines)),
	}

	for _, line := range lines {
		item := domain.CartItems{
			ProductId: line.ProductId,
			Quantity:  line.Quantity,
		}

		if s.catalog != nil {
			product, err := s.catalog.GetProduct(ctx, int64(line.ProductId))
			if err != nil {
				return nil, err
			}
			if product == nil || !product.Active {
				return nil, domain.ErrProductNotFound
			}
			item.Name = product.Name
			item.Price = product.Price
		}

		cart.Items = append(cart.Items, item)
		cart.Total += item.Price * float64(item.Quantity)
	}
	return cart, nil
}
func (s *CartService) validateProductAndStock(ctx context.Context, productID, qty int) error {
	if s.catalog == nil {
		return nil // MVP без catalog
	}
	product, err := s.catalog.GetProduct(ctx, int64(productID))
	if err != nil {
		return err
	}
	if product == nil || !product.Active {
		return domain.ErrProductNotFound
	}
	stock, err := s.catalog.GetAvailableStock(ctx, int64(productID))
	if err != nil {
		return err
	}
	if int32(qty) > stock {
		return domain.ErrInsufficientStock
	}
	return nil
}

func currentQty(items []domain.CartItems, productID int) int {
	for _, item := range items {
		if item.ProductId == productID {
			return item.Quantity
		}
	}
	return 0
}
