// PostgreSQL persistence layer for users.
// Слой хранения пользователей в PostgreSQL.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/repeter513/shop-auth/internal/domain"
)

// UserRepository stores and retrieves User records.
// UserRepository сохраняет и загружает записи User.
type UserRepository struct {
	// db is the shared connection pool; one pool per process.
	// db — общий пул соединений; один пул на процесс.
	db *pgxpool.Pool
}

// NewUserRepository creates a repository backed by a connection pool.
// NewUserRepository создаёт репозиторий на базе пула соединений.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser inserts a new user and fills generated fields on user.
// CreateUser вставляет нового пользователя и заполняет сгенерированные поля.
// Not idempotent: duplicate email returns unique-violation error from PostgreSQL.
// Не идемпотентно: повторный email возвращает ошибку unique-violation от PostgreSQL.
func (r *UserRepository) CreateUser(
	ctx context.Context,
	user *domain.User,
) error {
	// INSERT only email + hash; DB generates id, version, timestamps via RETURNING.
	// INSERT только email + hash; БД генерирует id, version, timestamps через RETURNING.
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

// GetUserByID loads a user by primary key.
// GetUserByID загружает пользователя по первичному ключу.
// Returns pgx.ErrNoRows when ID does not exist (caller maps to NotFound).
// Возвращает pgx.ErrNoRows если ID не существует (вызывающий код мапит в NotFound).
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

// GetUserByEmail loads a user by unique email address.
// GetUserByEmail загружает пользователя по уникальному email.
// Used by Login flow; ErrNoRows means unknown email (mapped to Unauthenticated, not NotFound).
// Используется в Login; ErrNoRows означает неизвестный email (мапится в Unauthenticated, не NotFound).
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
