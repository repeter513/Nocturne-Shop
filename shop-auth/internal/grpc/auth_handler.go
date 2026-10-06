// gRPC handlers for the AuthService protobuf contract.
// gRPC-обработчики контракта AuthService из protobuf.
package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-auth/internal/logx"
	"github.com/repeter513/shop-auth/internal/service"
	"github.com/repeter513/shop-auth/internal/validation"
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
	log  *slog.Logger
}

// NewHandler creates a gRPC handler backed by AuthService.
// NewHandler создаёт gRPC-обработчик на базе AuthService.
func NewHandler(auth *service.AuthService, log *slog.Logger) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{auth: auth, log: log}
}

// RegisterUser creates a new user account.
// RegisterUser создаёт новую учётную запись пользователя.
// Public RPC (no JWT). Duplicate email → AlreadyExists; other errors → Internal (details logged server-side).
// Публичный RPC (без JWT). Повторный email → AlreadyExists; прочие ошибки → Internal (детали только в логе).
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
		if errors.Is(err, validation.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		logx.LogHandlerErr(h.log, "RegisterUser", err)
		return nil, status.Error(codes.Internal, "registration failed")
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
		if errors.Is(err, validation.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		logx.LogHandlerErr(h.log, "LoginUser", err)
		return nil, status.Error(codes.Internal, "login failed")
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
		logx.LogHandlerErr(h.log, "GetUserInfo", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &authv1.GetUserInfoResponse{
		User: &authv1.User{Id: u.ID, Email: u.Email},
	}, nil
}
