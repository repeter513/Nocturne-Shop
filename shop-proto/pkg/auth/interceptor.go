package auth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(secret []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, err := userIDFromMetadata(ctx, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return handler(WithUserID(ctx, userID), req)
	}
}

func userIDFromMetadata(ctx context.Context, secret []byte) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return 0, errors.New("missing authorization")
	}
	token := strings.TrimSpace(vals[0])
	if token == "" {
		return 0, errors.New("missing authorization")
	}
	token, ok = strings.CutPrefix(token, "Bearer ")
	if !ok {
		return 0, errors.New("invalid authorization format")
	}
	userID, err := ParseAccessUserID(token, secret)
	if err != nil {
		return 0, errors.New("invalid token")
	}
	if userID == 0 {
		return 0, errors.New("invalid token")
	}
	return userID, nil
}
