// Domain models for the auth service.
// Доменные модели сервиса auth.
package domain

import "time"

// User represents a registered account in the system.
// User представляет зарегистрированную учётную запись в системе.
type User struct {
	// ID is the surrogate primary key assigned by the database (SERIAL).
	// ID — суррогатный первичный ключ, назначаемый БД (SERIAL).
	ID int64
	// Version supports optimistic locking; incremented on each update (default 1).
	// Version поддерживает optimistic locking; увеличивается при каждом обновлении (по умолчанию 1).
	Version int
	// Email is the unique login identifier (max 50 chars, UNIQUE constraint).
	// Email — уникальный идентификатор входа (макс. 50 символов, ограничение UNIQUE).
	Email string
	// PasswordHash stores bcrypt hash; never expose or log this field.
	// PasswordHash хранит bcrypt-хеш; никогда не отдавать наружу и не логировать.
	PasswordHash string
	// CreatedAt is set by DB DEFAULT CURRENT_TIMESTAMP on insert.
	// CreatedAt задаётся БД DEFAULT CURRENT_TIMESTAMP при вставке.
	CreatedAt time.Time
	// UpdatedAt is set by DB DEFAULT CURRENT_TIMESTAMP; bump on profile changes.
	// UpdatedAt задаётся БД DEFAULT CURRENT_TIMESTAMP; обновлять при изменении профиля.
	UpdatedAt time.Time
}
