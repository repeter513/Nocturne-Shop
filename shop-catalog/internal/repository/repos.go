// Package repository defines data access interfaces for the catalog service.
// Пакет repository определяет интерфейсы доступа к данным для сервиса каталога.
package repository

import (
	"context"

	"github.com/repeter513/shop-catalog/internal/domain"
)

// ProductRepository provides read access to products and categories.
// ProductRepository предоставляет чтение товаров и категорий.
type ProductRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, params ListProductsParams) ([]*domain.Product, int32, error)
	GetStock(ctx context.Context, productIDs []int64) (map[int64]int32, error)
	ListCategories(ctx context.Context, parentID *int64) ([]*domain.Category, error)
}

// ListProductsParams holds pagination and filter options for product listing.
// ListProductsParams содержит параметры пагинации и фильтрации списка товаров.
type ListProductsParams struct {
	// Page is 1-based page index; service defaults to 1 if <= 0.
	// Page — номер страницы с 1; service подставляет 1 если <= 0.
	Page int
	// PageSize is items per page; service defaults to 20, caps at 100.
	// PageSize — элементов на странице; service подставляет 20, максимум 100.
	PageSize int
	// CategoryID filters by category when > 0; 0 means all categories.
	// CategoryID фильтрует по категории если > 0; 0 — все категории.
	CategoryID int64
}
