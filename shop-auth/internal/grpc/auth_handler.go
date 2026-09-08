package grpc

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-auth/internal/service"
	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	auth *service.AuthService
}

func NewHandler(auth *service.AuthService) *Handler {
	return &Handler{auth: auth}
}
func (h *Handler) RegisterUser(
	ctx context.Context,
	req *authv1.RegisterUserRequest,
) (*authv1.RegisterUserResponse, error) {
	id, err := h.auth.Register(
		ctx,
		req.GetEmail(),
		req.GetPassword(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.RegisterUserResponse{UserId: id}, nil
}
func (h *Handler) LoginUser(
	ctx context.Context,
	req *authv1.LoginUserRequest,
) (*authv1.LoginUserResponse, error) {
	access, refresh, expiresIn, userID, err := h.auth.Login(
		ctx,
		req.GetEmail(),
		req.GetPassword(),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "invalid credentials" {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.LoginUserResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
		UserId:       userID,
	}, nil
}
func (h *Handler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, err := h.auth.Validate(
		ctx,
		req.GetToken(),
	)
	if err != nil {
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}
	return &authv1.ValidateTokenResponse{IsValid: true, UserId: userID}, nil
}
func (h *Handler) RefreshToken(
	ctx context.Context,
	req *authv1.RefreshTokenRequest,
) (*authv1.RefreshTokenResponse, error) {
	access, refresh, expiresIn, err := h.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	return &authv1.RefreshTokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
	}, nil
}
func (h *Handler) GetUserInfo(
	ctx context.Context,
	req *authv1.GetUserInfoRequest,
) (*authv1.GetUserInfoResponse, error) {
	u, err := h.auth.GetUser(
		ctx,
		req.GetUserId(),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.GetUserInfoResponse{
		User: &authv1.User{Id: u.ID, Email: u.Email},
	}, nil
}
