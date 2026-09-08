package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/repeter513/shop-auth/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	user *domain.User,
) error {
	return r.db.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
			 RETURNING id, version, created_at, updated_at`,
		user.Email,
		user.PasswordHash,
	).Scan(
		&user.ID,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

func (r *UserRepository) GetUserByID(
	ctx context.Context,
	id int,
	) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash, version, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash, version, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
