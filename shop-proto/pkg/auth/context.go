// Propagating authenticated user ID and roles through context.Context.
// Прокидывание user ID и ролей через context.Context.
package auth

import (
	"context"
	"errors"
)

// ctxKey is an unexported type to avoid context key collisions.
// ctxKey — неэкспортируемый тип для предотвращения коллизий ключей context.
type ctxKey struct{}

// authCtx holds the authenticated principal extracted from JWT.
// authCtx хранит аутентифицированного пользователя, извлечённого из JWT.
type authCtx struct {
	userID int64
	roles  []string
}

// ErrUserIDMissing is returned when a handler expects auth but context has no user.
// ErrUserIDMissing возвращается, когда handler ожидает auth, но user ID отсутствует в context.
var ErrUserIDMissing = errors.New("user id missing from context")

// WithUserID attaches user ID to context (no roles).
// WithUserID добавляет user ID в context (без ролей).
func WithUserID(ctx context.Context, userID int64) context.Context {
	return WithAuth(ctx, userID, nil)
}

// WithAuth attaches user ID and roles after JWT validation in interceptor.
// WithAuth добавляет user ID и роли после проверки JWT в interceptor.
func WithAuth(ctx context.Context, userID int64, roles []string) context.Context {
	return context.WithValue(ctx, ctxKey{}, authCtx{userID: userID, roles: roles})
}

// UserIDFromContext extracts the authenticated user ID set by the interceptor.
// UserIDFromContext извлекает user ID, установленный interceptor.
func UserIDFromContext(ctx context.Context) (int64, error) {
	switch v := ctx.Value(ctxKey{}).(type) {
	case authCtx:
		if v.userID == 0 {
			return 0, ErrUserIDMissing
		}
		return v.userID, nil
	case int64:
		// Legacy path: bare int64 stored directly (backward compat).
		// Legacy-путь: int64 сохранён напрямую (обратная совместимость).
		if v == 0 {
			return 0, ErrUserIDMissing
		}
		return v, nil
	default:
		return 0, ErrUserIDMissing
	}
}

// RolesFromContext returns RBAC roles from context, or nil if absent.
// RolesFromContext возвращает RBAC-роли из context или nil, если отсутствуют.
func RolesFromContext(ctx context.Context) []string {
	if v, ok := ctx.Value(ctxKey{}).(authCtx); ok {
		return v.roles
	}
	return nil
}

// HasRole checks whether the context contains the given role name.
// HasRole проверяет, содержит ли context указанную роль.
func HasRole(ctx context.Context, role string) bool {
	for _, r := range RolesFromContext(ctx) {
		if r == role {
			return true
		}
	}
	return false
}
