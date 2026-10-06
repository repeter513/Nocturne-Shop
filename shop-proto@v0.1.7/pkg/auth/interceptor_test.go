// Tests for gRPC UnaryServerInterceptor: public skip, auth required, role propagation.
// Тесты gRPC UnaryServerInterceptor: пропуск public, требование auth, проброс ролей.
package auth

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewSigner(priv, time.Hour, time.Hour, Issuer, AccessAudiences)
	verifier := NewVerifier(pub, Issuer, AudienceAuth)

	tok, err := signer.IssueAccessToken(42, []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	ic := UnaryServerInterceptor(verifier, PublicAuth...)
	passthrough := func(ctx context.Context, _ any) (any, error) {
		id, err := UserIDFromContext(ctx)
		if err != nil {
			return int64(0), nil
		}
		return id, nil
	}

	// Public method (RegisterUser) must not require a token.
	// Публичный метод (RegisterUser) не должен требовать токен.
	got, err := ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: PublicAuth[0]}, passthrough)
	if err != nil {
		t.Fatal(err)
	}
	if got.(int64) != 0 {
		t.Fatalf("public method must skip auth, got user %v", got)
	}

	// Protected method without token must return Unauthenticated.
	// Защищённый метод без токена должен вернуть Unauthenticated.
	_, err = ic(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/GetUserInfo"}, passthrough)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("protected without token: %v", err)
	}

	md := metadata.Pairs("authorization", "Bearer "+tok)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	got, err = ic(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/GetUserInfo"}, passthrough)
	if err != nil {
		t.Fatal(err)
	}
	if got.(int64) != 42 {
		t.Fatalf("want user 42, got %v", got)
	}

	// Roles from JWT must be available in handler context.
	// Роли из JWT должны быть доступны в context handler.
	roleCheck := func(ctx context.Context, _ any) (any, error) {
		return HasRole(ctx, "admin"), nil
	}
	got, err = ic(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/GetUserInfo"}, roleCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !got.(bool) {
		t.Fatal("want admin role in context")
	}
}
