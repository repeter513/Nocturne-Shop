// Package http implements REST handlers and middleware for the BFF.
// Пакет http реализует REST-обработчики и middleware для BFF.
package http

import (
	"net/http"
	"strings"
)

// cors wraps a handler with CORS headers for the given origins.
// cors оборачивает обработчик CORS-заголовками для указанных origins.
//
// Applied to every /api/v1/* route so the React dev server (port 3000)
// can call the BFF (port 8090) cross-origin during development.
// Применяется ко всем маршрутам /api/v1/*, чтобы React dev-сервер (порт 3000)
// мог обращаться к BFF (порт 8090) cross-origin при разработке.
func cors(origins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow browser clients from the configured origin(s).
			// Разрешаем браузерным клиентам с указанного origin.
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			// Authorization header is required for JWT bearer tokens from the SPA.
			// Заголовок Authorization нужен для JWT bearer-токенов из SPA.
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			// Preflight OPTIONS — respond 204 without hitting the handler.
			// Preflight OPTIONS — ответ 204 без вызова обработчика.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// requireAuth rejects requests without a Bearer token.
// requireAuth отклоняет запросы без Bearer-токена.
//
// This is a lightweight gate: it only checks header presence/format.
// Full JWT validation happens in the auth gRPC service (ValidateToken).
// Это лёгкий шлюз: проверяется только наличие/формат заголовка.
// Полная валидация JWT выполняется в gRPC-сервисе auth (ValidateToken).
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization"})
			return
		}
		next(w, r)
	}
}
