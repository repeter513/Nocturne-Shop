// gRPC server interceptor for JWT authentication on protected methods.
// gRPC server interceptor для JWT-аутентификации на защищённых методах.
package auth

import (
	"context"
	"errors"
	"strings"

	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// PublicAuth lists auth-service methods that skip JWT validation.
// PublicAuth — методы auth-сервиса без проверки JWT.
var PublicAuth = []string{
	authv1.AuthService_RegisterUser_FullMethodName,
	authv1.AuthService_LoginUser_FullMethodName,
	authv1.AuthService_ValidateToken_FullMethodName,
	authv1.AuthService_RefreshToken_FullMethodName,
}

// PublicCatalog lists catalog read methods available without authentication.
// PublicCatalog — методы чтения каталога, доступные без аутентификации.
var PublicCatalog = []string{
	catalogv1.CatalogService_GetProduct_FullMethodName,
	catalogv1.CatalogService_ListProducts_FullMethodName,
	catalogv1.CatalogService_ListCategories_FullMethodName,
	catalogv1.CatalogService_GetStock_FullMethodName,
}

// UnaryServerInterceptor validates Bearer tokens and injects user ID into context.
// UnaryServerInterceptor проверяет Bearer-токены и добавляет user ID в context.
// Pass method full names as public to skip auth (e.g. login, register).
// Передайте full method name как public для пропуска auth (login, register).
func UnaryServerInterceptor(v *Verifier, public ...string) grpc.UnaryServerInterceptor {
	skip := make(map[string]struct{}, len(public))
	for _, m := range public {
		skip[m] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Public endpoints bypass token check entirely.
		// Публичные эндпоинты полностью пропускают проверку токена.
		if _, ok := skip[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		claims, err := claimsFromMetadata(ctx, v)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return handler(WithAuth(ctx, claims.UserID, claims.Roles), req)
	}
}

// claimsFromMetadata extracts and validates the Bearer token from gRPC metadata.
// claimsFromMetadata извлекает и проверяет Bearer-токен из gRPC metadata.
func claimsFromMetadata(ctx context.Context, v *Verifier) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return nil, errors.New("missing authorization")
	}
	token := strings.TrimSpace(vals[0])
	if token == "" {
		return nil, errors.New("missing authorization")
	}
	// Expect "Bearer <token>" format per RFC 6750.
	// Ожидаем формат "Bearer <token>" по RFC 6750.
	token, ok = strings.CutPrefix(token, "Bearer ")
	if !ok {
		return nil, errors.New("invalid authorization format")
	}
	claims, err := v.ParseAccess(token)
	if err != nil || claims.UserID == 0 {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
