package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/repeter513/shop-auth/internal/config"
	authgrpc "github.com/repeter513/shop-auth/internal/grpc"
	"github.com/repeter513/shop-auth/internal/repository"
	"github.com/repeter513/shop-auth/internal/service"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	db, err := pgxpool.New(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("db ping:", err)
	}
	tokens := pkgauth.NewJWT(
		[]byte(cfg.JWTSecret),
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	repo := repository.NewUserRepository(db)
	auth := service.NewAuthService(repo, tokens)
	handler := authgrpc.NewHandler(auth)
	lis, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(srv, handler)
	reflection.Register(srv)
	log.Println("gRPC listen", cfg.GRPCAddr())
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
