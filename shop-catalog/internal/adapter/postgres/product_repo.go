package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-catalog/internal/domain"
	"github.com/repeter513/shop-catalog/internal/repository"
)

type ProductRepo struct {
	db *DB
}

func NewProductRepo(db *DB) *ProductRepo {
	return &ProductRepo{db: db}
}

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

func (r *ProductRepo) getStockForUpdate(ctx context.Context, tx pgx.Tx, productID int64) (int32, error) {
	var stock int32
	err := tx.QueryRow(ctx, `SELECT stock FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&stock)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrProductNotFound
	}
	return stock, err
}

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

type rowScanner interface {
	Scan(dest ...any) error
}

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
