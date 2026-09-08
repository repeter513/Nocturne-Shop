package repository

import (
	"context"

	"github.com/repeter513/shop-catalog/internal/domain"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, params ListProductsParams) ([]*domain.Product, int32, error)
	GetStock(ctx context.Context, productIDs []int64) (map[int64]int32, error)
	ListCategories(ctx context.Context, parentID *int64) ([]*domain.Category, error)
}

type ListProductsParams struct {
	Page       int
	PageSize   int
	CategoryID int64
}
