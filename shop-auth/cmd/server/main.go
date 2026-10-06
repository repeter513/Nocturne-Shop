// Auth microservice entry point: gRPC server for registration, login, and JWT.
// Точка входа микросервиса auth: gRPC-сервер регистрации, входа и JWT.
package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/repeter513/shop-auth/internal/config"
	authgrpc "github.com/repeter513/shop-auth/internal/grpc"
	"github.com/repeter513/shop-auth/internal/logx"
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
	logger := logx.New(cfg.LogLevel)
	slog.SetDefault(logger)
	ctx := context.Background()

	// Step 2: open PostgreSQL pool and verify connectivity before accepting traffic.
	// Шаг 2: открытие пула PostgreSQL и проверка соединения до приёма запросов.
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		logger.Error("db ping", slog.Any("err", err))
		os.Exit(1)
	}

	// Step 3: JWT signer issues access/refresh tokens; verifier validates incoming RPC auth.
	// Шаг 3: signer выдаёт access/refresh токены; verifier проверяет JWT во входящих RPC.
	// Stateless JWT only: access short TTL; refresh re-mints a new pair without server-side revocation (old refresh valid until exp).
	// Только stateless JWT: короткий access; refresh выдаёт новую пару без отзыва на сервере (старый refresh действует до exp).
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
	handler := authgrpc.NewHandler(auth, logger)

	// Step 5: bind TCP listener on GRPC_PORT (default from env, e.g. :8081).
	// Шаг 5: привязка TCP-слушателя к GRPC_PORT (из env, например :8081).
	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		logger.Error("listen", slog.String("addr", cfg.GRPCAddr()), slog.Any("err", err))
		os.Exit(1)
	}

	// Step 6: unary interceptor validates JWT on protected methods; PublicAuth lists RPCs that skip auth (Register, Login, Validate, Refresh).
	// Шаг 6: unary interceptor проверяет JWT на защищённых методах; PublicAuth — RPC без auth (Register, Login, Validate, Refresh).
	srv := grpc.NewServer(grpc.UnaryInterceptor(pkgauth.UnaryServerInterceptor(verifier, pkgauth.PublicAuth...)))
	authv1.RegisterAuthServiceServer(srv, handler)
	// Reflection enables grpcurl/grpcui introspection in dev; safe to expose only behind internal network.
	// Reflection включает introspection для grpcurl/grpcui в dev; безопасно только во внутренней сети.
	reflection.Register(srv)

	go func() {
		logger.Info("gRPC listen", slog.String("addr", cfg.GRPCAddr()))
		if err := srv.Serve(lis); err != nil {
			logger.Error("gRPC serve", slog.Any("err", err))
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	srv.GracefulStop()
}
