// Package service implements cart business logic.
// Пакет service реализует бизнес-логику корзины.
package service

import (
	"context"
	"github.com/repeter513/shop-cart/internal/adapter/catalog"
	"github.com/repeter513/shop-cart/internal/domain"
	"github.com/repeter513/shop-cart/internal/repository"
)

// CartService orchestrates cart operations with persistence and catalog validation.
// CartService координирует операции с корзиной, хранилищем и проверкой каталога.
type CartService struct {
	// repo persists cart line items per user in PostgreSQL.
	// repo сохраняет позиции корзины пользователя в PostgreSQL.
	repo repository.CartItemRepository
	// catalog is the downstream gRPC client for product and stock lookups.
	// catalog — downstream gRPC-клиент для запросов товаров и остатков.
	catalog *catalog.Client
}

// NewCartService creates a CartService with the given repository and catalog client.
// NewCartService создаёт CartService с указанным репозиторием и клиентом каталога.
func NewCartService(
	repo repository.CartItemRepository,
	catalog *catalog.Client) *CartService {
	return &CartService{
		repo:    repo,
		catalog: catalog,
	}
}

// GetCart returns the current cart for the given user.
// GetCart возвращает текущую корзину указанного пользователя.
func (s *CartService) GetCart(ctx context.Context, userID int) (*domain.Cart, error) {
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	return s.buildCart(ctx, userID)
}

// AddToCart adds a product to the cart or increases its quantity.
// AddToCart добавляет товар в корзину или увеличивает его количество.
func (s *CartService) AddToCart(ctx context.Context, userID, productID int, quantity int32) (*domain.Cart, error) {
	// Step 1: validate input identifiers and quantity.
	// Шаг 1: проверка входных идентификаторов и количества.
	if userID == 0 {
		return nil, domain.ErrIDRequired
	}
	if productID == 0 {
		return nil, domain.ErrProductRequired
	}
	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	// Step 2: load existing cart lines to compute the new total quantity.
	// Шаг 2: загрузка текущих позиций для вычисления нового итогового количества.
	items, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	newQty := currentQty(items, productID) + int(quantity)

	// Step 3: verify product exists and stock covers the requested quantity via catalog.
	// Шаг 3: проверка существования товара и достаточности остатка через каталог.
	if err := s.validateProductAndStock(ctx, productID, newQty); err != nil {
		return nil, err
	}

	// Step 4: upsert the line item (insert or update quantity on conflict).
	// Шаг 4: upsert позиции (вставка или обновление количества при конфликте).
	if err := s.repo.UpsertItems(ctx, userID, productID, int32(newQty)); err != nil {
		return nil, err
	}

	// Step 5: rebuild and return the enriched cart with catalog prices.
	// Шаг 5: пересборка и возврат обогащённой корзины с ценами из каталога.
	return s.buildCart(ctx, userID)
}

// UpdateCartItem sets the quantity of a cart item; zero removes the item.
// UpdateCartItem устанавливает количество позиции; ноль удаляет её из корзины.
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
	// Zero quantity is treated as a remove operation.
	// Нулевое количество трактуется как операция удаления.
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

// RemoveFromCart deletes a product from the user's cart.
// RemoveFromCart удаляет товар из корзины пользователя.
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

// ClearCart removes all items from the user's cart.
// ClearCart удаляет все позиции из корзины пользователя.
func (s *CartService) ClearCart(ctx context.Context, userID int) error {
	if userID == 0 {
		return domain.ErrIDRequired
	}
	return s.repo.Clear(ctx, userID)
}

// buildCart loads cart lines and enriches them with catalog data and totals.
// buildCart загружает позиции корзины, обогащает данными каталога и считает итог.
func (s *CartService) buildCart(ctx context.Context, userID int) (*domain.Cart, error) {
	// Step 1: fetch raw cart lines from PostgreSQL (product_id + quantity only).
	// Шаг 1: загрузка сырых позиций из PostgreSQL (только product_id + quantity).
	lines, err := s.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	cart := &domain.Cart{
		UserId: userID,
		Items:  make([]domain.CartItems, 0, len(lines)),
	}

	// Step 2: for each line, call catalog.GetProduct to enrich name and price.
	// Шаг 2: для каждой позиции вызов catalog.GetProduct для обогащения имени и цены.
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
		// Step 3: accumulate running total from enriched prices.
		// Шаг 3: накопление итоговой суммы из обогащённых цен.
		cart.Total += item.Price * int64(item.Quantity)
	}
	return cart, nil
}

// validateProductAndStock checks that the product exists and stock is sufficient.
// validateProductAndStock проверяет наличие товара и достаточность остатка.
func (s *CartService) validateProductAndStock(ctx context.Context, productID, qty int) error {
	if s.catalog == nil {
		return nil // MVP without catalog / MVP без каталога
	}
	// Step 1: downstream call — catalog.GetProduct to verify product is active.
	// Шаг 1: downstream-вызов — catalog.GetProduct для проверки активности товара.
	product, err := s.catalog.GetProduct(ctx, int64(productID))
	if err != nil {
		return err
	}
	if product == nil || !product.Active {
		return domain.ErrProductNotFound
	}
	// Step 2: downstream call — catalog.GetAvailableStock to check inventory.
	// Шаг 2: downstream-вызов — catalog.GetAvailableStock для проверки остатка.
	stock, err := s.catalog.GetAvailableStock(ctx, int64(productID))
	if err != nil {
		return err
	}
	if int32(qty) > stock {
		return domain.ErrInsufficientStock
	}
	return nil
}

// currentQty returns the existing quantity of a product in the cart.
// currentQty возвращает текущее количество товара в корзине.
func currentQty(items []domain.CartItems, productID int) int {
	for _, item := range items {
		if item.ProductId == productID {
			return item.Quantity
		}
	}
	return 0
}
