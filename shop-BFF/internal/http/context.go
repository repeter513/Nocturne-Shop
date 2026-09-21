// Package http implements REST handlers and middleware for the BFF.
// Пакет http реализует REST-обработчики и middleware для BFF.
package http

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// withAuth forwards the Authorization header into outgoing gRPC metadata.
// withAuth передаёт заголовок Authorization в исходящие gRPC-метаданные.
//
// Backend services read metadata["authorization"] to identify the caller
// and validate the JWT on protected RPCs (cart, order, payment).
// Backend-сервисы читают metadata["authorization"] для идентификации
// вызывающего и валидации JWT на защищённых RPC (cart, order, payment).
func withAuth(ctx context.Context, authHeader string) context.Context {
	if authHeader == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", authHeader)
}
