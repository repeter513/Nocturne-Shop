package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-auth/internal/domain"
	"github.com/repeter513/shop-auth/internal/repository"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *pkgauth.JWT
}

func NewAuthService(
	users *repository.UserRepository,
	tokens *pkgauth.JWT,
) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email,
	password string,
) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.CreateUser(ctx, u); err != nil {
		return 0, err
	}
	return u.ID, nil
}
func (s *AuthService) Login(
	ctx context.Context,
	email,
	password string,
) (
	access,
	refresh string,
	expiresIn,
	userID int64,
	err error,
) {
	u, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", "", 0, 0, errors.New("invalid credentials")
	}
	access, err = s.tokens.IssueAccessToken(u.ID)
	if err != nil {
		return "", "", 0, 0, err
	}
	refresh, err = s.tokens.IssueRefreshToken(u.ID)
	if err != nil {
		return "", "", 0, 0, err
	}
	return access, refresh, s.tokens.AccessTTLSeconds(), u.ID, nil
}
func (s *AuthService) Validate(
	ctx context.Context,
	accessToken string,
) (userID int64, err error) {
	c, err := s.tokens.ParseAccess(accessToken)
	if err != nil {
		return 0, err
	}
	return c.UserID, nil
}
func (s *AuthService) Refresh(
	_ context.Context,
	refreshToken string,
) (
	access,
	refresh string,
	expiresIn int64,
	err error,
) {
	c, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", 0, err
	}
	access, err = s.tokens.IssueAccessToken(c.UserID)
	if err != nil {
		return "", "", 0, err
	}
	refresh, err = s.tokens.IssueRefreshToken(c.UserID)
	if err != nil {
		return "", "", 0, err
	}
	return access, refresh, s.tokens.AccessTTLSeconds(), nil
}
func (s *AuthService) GetUser(
	ctx context.Context,
	id int64,
) (*domain.User, error) {
	u, err := s.users.GetUserByID(ctx, int(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return u, err
}
