package http

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func withAuth(ctx context.Context, authHeader string) context.Context {
	if authHeader == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", authHeader)
}
