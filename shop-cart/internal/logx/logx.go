// Package logx configures JSON slog from LOG_LEVEL.
// Пакет logx настраивает JSON slog по LOG_LEVEL.
package logx

import (
	"log/slog"
	"os"
	"strings"
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
