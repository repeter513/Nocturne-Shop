package auth

import (
	"context"
	"errors"
)

type ctxKey struct{}

var ErrUserIDMissing = errors.New("user id missing from context")

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int64, error) {
	v, ok := ctx.Value(ctxKey{}).(int64)
	if !ok || v == 0 {
		return 0, ErrUserIDMissing
	}
	return v, nil
}
