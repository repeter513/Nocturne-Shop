// gRPC handlers for the AuthService protobuf contract.
// gRPC-обработчики контракта AuthService из protobuf.
package grpc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-auth/internal/service"
	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements authv1.AuthServiceServer.
// Handler реализует authv1.AuthServiceServer.
type Handler struct {
	authv1.UnimplementedAuthServiceServer
	// auth contains registration, login, JWT validate/refresh, and profile logic.
	// auth содержит логику регистрации, входа, validate/refresh JWT и профиля.
	auth *service.AuthService
}

// NewHandler creates a gRPC handler backed by AuthService.
// NewHandler создаёт gRPC-обработчик на базе AuthService.
func NewHandler(auth *service.AuthService) *Handler {
	return &Handler{auth: auth}
}

// RegisterUser creates a new user account.
// RegisterUser создаёт новую учётную запись пользователя.
// Public RPC (no JWT). Duplicate email → Internal (DB unique violation), not AlreadyExists.
// Публичный RPC (без JWT). Повторный email → Internal (unique violation в БД), не AlreadyExists.
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

// LoginUser authenticates credentials and returns JWT tokens.
// LoginUser проверяет учётные данные и возвращает JWT-токены.
// Public RPC. Returns access + refresh + expires_in (seconds) + user_id.
// Публичный RPC. Возвращает access + refresh + expires_in (секунды) + user_id.
// Error mapping: unknown email or bad password → Unauthenticated (no user enumeration).
// Маппинг ошибок: неизвестный email или неверный пароль → Unauthenticated (без перечисления пользователей).
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

// ValidateToken checks whether an access token is valid.
// ValidateToken проверяет, действителен ли access-token.
// Public RPC for inter-service auth checks. Invalid/expired token → IsValid=false (not an error).
// Публичный RPC для межсервисных проверок. Невалидный/просроченный токен → IsValid=false (не ошибка).
func (h *Handler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, roles, err := h.auth.Validate(ctx, req.GetToken())
	if err != nil {
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}
	return &authv1.ValidateTokenResponse{IsValid: true, UserId: userID, Roles: roles}, nil
}

// RefreshToken exchanges a refresh token for a new token pair.
// RefreshToken обменивает refresh-токен на новую пару токенов.
// Public RPC. Invalid refresh → Unauthenticated. Successful refresh rotates both tokens.
// Публичный RPC. Невалидный refresh → Unauthenticated. Успешный refresh ротирует оба токена.
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

// GetUserInfo returns profile data for the authenticated user.
// GetUserInfo возвращает данные профиля аутентифицированного пользователя.
// Protected RPC: JWT extracted from context by UnaryServerInterceptor (Bearer metadata).
// Защищённый RPC: JWT извлекается из context через UnaryServerInterceptor (Bearer metadata).
// Error paths: missing JWT → Unauthenticated; deleted user → NotFound; DB error → Internal.
// Пути ошибок: нет JWT → Unauthenticated; удалённый user → NotFound; ошибка БД → Internal.
func (h *Handler) GetUserInfo(
	ctx context.Context,
	_ *authv1.GetUserInfoRequest,
) (*authv1.GetUserInfoResponse, error) {
	userID, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	u, err := h.auth.GetUser(ctx, userID)
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
