// Package logx configures JSON slog and RPC error logging helpers.
// Пакет logx настраивает JSON slog и хелперы логирования ошибок RPC.
package logx

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-auth/internal/service"
)

func New(level string) *slog.Logger {
	var min slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		min = slog.LevelDebug
	case "warn", "warning":
		min = slog.LevelWarn
	case "error":
		min = slog.LevelError
	default:
		min = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: min}))
}

func LevelForHandlerErr(err error) slog.Level {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists),
		errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, pgx.ErrNoRows):
		return slog.LevelDebug
	default:
		return slog.LevelError
	}
}

func LogHandlerErr(l *slog.Logger, rpc string, err error) {
	if err == nil {
		return
	}
	lvl := LevelForHandlerErr(err)
	if !l.Enabled(context.Background(), lvl) {
		return
	}
	l.Log(context.Background(), lvl, "rpc error", slog.String("rpc", rpc), slog.Any("err", err))
}
