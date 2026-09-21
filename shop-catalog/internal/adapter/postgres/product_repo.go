// Package postgres implements product and category persistence.
// Пакет postgres реализует хранение товаров и категорий.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
)

// ProductRepo provides PostgreSQL-backed product and category access.
// ProductRepo предоставляет доступ к товарам и категориям через PostgreSQL.
type ProductRepo struct {
	// db is the shared connection wrapper with pool and WithTx helper.
	// db — общая обёртка соединения с pool и WithTx.
	db *DB
}

// NewProductRepo creates a ProductRepo bound to the given DB.
// NewProductRepo создаёт ProductRepo, привязанный к указанной БД.
func NewProductRepo(db *DB) *ProductRepo {
	return &ProductRepo{db: db}
}

// FindByID loads a product by primary key; nil if not found.
// FindByID загружает товар по первичному ключу; nil если не найден.
func (r *ProductRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	row := r.db.Pool().QueryRow(ctx, `
		SELECT id, name, description, price, category_id, stock, active, created_at, updated_at
		FROM products WHERE id = $1`, id)
	p, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// List returns active products with pagination and optional category filter.
// List возвращает активные товары с пагинацией и опциональным фильтром по категории.
// SQL: COUNT for total, then SELECT ... ORDER BY id LIMIT/OFFSET (params from service defaults).
// SQL: COUNT для total, затем SELECT ... ORDER BY id LIMIT/OFFSET (параметры из дефолтов service).
func (r *ProductRepo) List(ctx context.Context, params repository.ListProductsParams) ([]*domain.Product, int32, error) {
	where := "WHERE active = true"
	args := []any{}
	argN := 1

	if params.CategoryID > 0 {
		where += fmt.Sprintf(" AND category_id = $%d", argN)
		args = append(args, params.CategoryID)
		argN++
	}

	var total int32
	if err := r.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM products "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf(`
		SELECT id, name, description, price, category_id, stock, active, created_at, updated_at
		FROM products %s ORDER BY id LIMIT $%d OFFSET $%d`, where, argN, argN+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

// GetStock returns physical stock quantities for the given product IDs.
// GetStock возвращает физические остатки для указанных ID товаров.
// Does not subtract reservations — service layer computes available stock.
// Не вычитает резервы — service слой вычисляет доступный остаток.
func (r *ProductRepo) GetStock(ctx context.Context, productIDs []int64) (map[int64]int32, error) {
	rows, err := r.db.Pool().Query(ctx, `SELECT id, stock FROM products WHERE id = ANY($1)`, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]int32, len(productIDs))
	for rows.Next() {
		var id int64
		var stock int32
		if err := rows.Scan(&id, &stock); err != nil {
			return nil, err
		}
		out[id] = stock
	}
	return out, rows.Err()
}

// ListCategories returns categories, optionally filtered by parent ID.
// ListCategories возвращает категории, опционально отфильтрованные по parent ID.
func (r *ProductRepo) ListCategories(ctx context.Context, parentID *int64) ([]*domain.Category, error) {
	var rows pgx.Rows
	var err error

	if parentID == nil {
		rows, err = r.db.Pool().Query(ctx, `
			SELECT id, name, description, parent_id, created_at FROM categories ORDER BY id`)
	} else {
		rows, err = r.db.Pool().Query(ctx, `
			SELECT id, name, description, parent_id, created_at FROM categories WHERE parent_id = $1 ORDER BY id`, *parentID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, rows.Err()
}

// getStockForUpdate locks a product row and returns its stock for transactional updates.
// getStockForUpdate блокирует строку товара и возвращает остаток для транзакционных обновлений.
// SELECT ... FOR UPDATE prevents concurrent reserve/confirm races on same product.
// SELECT ... FOR UPDATE предотвращает гонки reserve/confirm на одном товаре.
func (r *ProductRepo) getStockForUpdate(ctx context.Context, tx pgx.Tx, productID int64) (int32, error) {
	var stock int32
	err := tx.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrProductNotFound
	}
	return stock, err
}

// decrementStockTx atomically reduces product stock if sufficient quantity exists.
// decrementStockTx атомарно уменьшает остаток товара при достаточном количестве.
// RowsAffected=0 → ErrInsufficientStock (optimistic check via WHERE stock >= qty).
// RowsAffected=0 → ErrInsufficientStock (оптимистичная проверка через WHERE stock >= qty).
func (r *ProductRepo) decrementStockTx(ctx context.Context, tx pgx.Tx, productID int64, qty int32) error {
	tag, err := tx.Exec(ctx, `
		UPDATE products SET stock = stock - $2, updated_at = NOW()
		WHERE id = $1 AND stock >= $2`, productID, qty)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

// rowScanner abstracts pgx.Row and pgx.Rows Scan method.
// rowScanner абстрагирует метод Scan pgx.Row и pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanProduct maps a database row to a domain Product.
// scanProduct преобразует строку БД в доменный Product.
func scanProduct(row rowScanner) (*domain.Product, error) {
	var p domain.Product
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.CategoryID,
		&p.Stock,
		&p.Active,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
