// Auth microservice entry point: gRPC server for registration, login, and JWT.
// Точка входа микросервиса auth: gRPC-сервер регистрации, входа и JWT.
package main

import (
	"context"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/repeter513/shop-auth/internal/config"
	authgrpc "github.com/repeter513/shop-auth/internal/grpc"
	"github.com/repeter513/shop-auth/internal/repository"
	"github.com/repeter513/shop-auth/internal/service"
	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Step 1: load config from .env and required env vars; fail fast on missing JWT keys or DB URL.
	// Шаг 1: загрузка конфига из .env и обязательных переменных; немедленный выход при отсутствии JWT-ключей или DATABASE_URL.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Step 2: open PostgreSQL pool and verify connectivity before accepting traffic.
	// Шаг 2: открытие пула PostgreSQL и проверка соединения до приёма запросов.
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("db ping:", err)
	}

	// Step 3: JWT signer issues access/refresh tokens; verifier validates incoming RPC auth.
	// Шаг 3: signer выдаёт access/refresh токены; verifier проверяет JWT во входящих RPC.
	// Access tokens carry short TTL; refresh tokens enable rotation without re-login (not idempotent — each refresh mints new pair).
	// Access-токены с коротким TTL; refresh-токены позволяют ротацию без повторного входа (не идемпотентно — каждый refresh выдаёт новую пару).
	signer := pkgauth.NewSigner(
		cfg.JWTPrivateKey,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
		pkgauth.Issuer,
		pkgauth.AccessAudiences,
	)
	verifier := pkgauth.NewVerifier(cfg.JWTPublicKey, pkgauth.Issuer, pkgauth.AudienceAuth)

	// Step 4: wire repository → service → gRPC handler.
	// Шаг 4: сборка цепочки repository → service → gRPC handler.
	repo := repository.NewUserRepository(db)
	auth := service.NewAuthService(repo, signer, verifier)
	handler := authgrpc.NewHandler(auth)

	// Step 5: bind TCP listener on GRPC_PORT (default from env, e.g. :8081).
	// Шаг 5: привязка TCP-слушателя к GRPC_PORT (из env, например :8081).
	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}

	// Step 6: unary interceptor validates JWT on protected methods; PublicAuth lists RPCs that skip auth (Register, Login, Validate, Refresh).
	// Шаг 6: unary interceptor проверяет JWT на защищённых методах; PublicAuth — RPC без auth (Register, Login, Validate, Refresh).
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier, pkgauth.PublicAuth...)))
	authv1.RegisterAuthServiceServer(srv, handler)
	// Reflection enables grpcurl/grpcui introspection in dev; safe to expose only behind internal network.
	// Reflection включает introspection для grpcurl/grpcui в dev; безопасно только во внутренней сети.
	reflection.Register(srv)

	log.Println("gRPC listen", cfg.GRPCAddr())
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
