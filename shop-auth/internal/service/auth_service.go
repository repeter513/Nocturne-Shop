// Business logic for registration, login, and token lifecycle.
// Бизнес-логика регистрации, входа и жизненного цикла токенов.
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

// AuthService orchestrates user auth and JWT operations.
// AuthService координирует аутентификацию пользователей и операции с JWT.
type AuthService struct {
	// users persists and loads User entities.
	// users сохраняет и загружает сущности User.
	users *repository.UserRepository
	// signer mints signed JWT access and refresh tokens with configured TTLs.
	// signer выпускает подписанные JWT access и refresh токены с настроенными TTL.
	signer *pkgauth.Signer
	// verifier parses and validates JWT signatures and claims (issuer, audience, expiry).
	// verifier разбирает и проверяет подписи и claims JWT (issuer, audience, expiry).
	verifier *pkgauth.Verifier
}

// NewAuthService wires repository and JWT helpers.
// NewAuthService связывает репозиторий и JWT-хелперы.
func NewAuthService(
	users *repository.UserRepository,
	signer *pkgauth.Signer,
	verifier *pkgauth.Verifier,
) *AuthService {
	return &AuthService{
		users:    users,
		signer:   signer,
		verifier: verifier,
	}
}

// Register hashes the password and persists a new user.
// Register хеширует пароль и сохраняет нового пользователя.
// Idempotency: none — duplicate email fails at DB unique constraint.
// Идемпотентность: нет — повторный email падает на unique constraint в БД.
func (s *AuthService) Register(
	ctx context.Context,
	email,
	password string,
) (int64, error) {
	// bcrypt.DefaultCost (10) balances security and latency on login.
	// bcrypt.DefaultCost (10) балансирует безопасность и задержку при входе.
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

// Login verifies credentials and issues access/refresh tokens.
// Login проверяет учётные данные и выдаёт access/refresh токены.
// JWT flow: lookup user → bcrypt compare → IssueAccessToken (short TTL) → IssueRefreshToken (long TTL).
// JWT flow: поиск user → bcrypt compare → IssueAccessToken (короткий TTL) → IssueRefreshToken (длинный TTL).
// Error paths: ErrNoRows or wrong password both surface as "invalid credentials" at handler layer.
// Пути ошибок: ErrNoRows или неверный пароль на уровне handler оба дают "invalid credentials".
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
	// Access token: no roles on first login (nil roles slice).
	// Access token: без ролей при первом входе (nil roles slice).
	access, err = s.signer.IssueAccessToken(u.ID, nil)
	if err != nil {
		return "", "", 0, 0, err
	}
	refresh, _, err = s.signer.IssueRefreshToken(u.ID)
	if err != nil {
		return "", "", 0, 0, err
	}
	return access, refresh, s.signer.AccessTTLSeconds(), u.ID, nil
}

// Validate parses and verifies an access token.
// Validate разбирает и проверяет access-токен.
// Idempotent read: same token returns same userID/roles until expiry or revocation.
// Идемпотентное чтение: один токен возвращает те же userID/roles до истечения или отзыва.
func (s *AuthService) Validate(
	ctx context.Context,
	accessToken string,
) (userID int64, roles []string, err error) {
	c, err := s.verifier.ParseAccess(accessToken)
	if err != nil {
		return 0, nil, err
	}
	return c.UserID, c.Roles, nil
}

// Refresh validates a refresh token and issues a new token pair.
// Refresh проверяет refresh-токен и выдаёт новую пару токенов.
// Not idempotent: each call rotates tokens; old refresh may be invalidated by signer policy.
// Не идемпотентно: каждый вызов ротирует токены; старый refresh может быть инвалидирован политикой signer.
// Preserves roles from refresh claims into new access token.
// Сохраняет roles из refresh claims в новый access token.
func (s *AuthService) Refresh(
	_ context.Context,
	refreshToken string,
) (
	access,
	refresh string,
	expiresIn int64,
	err error,
) {
	c, err := s.verifier.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", 0, err
	}
	access, err = s.signer.IssueAccessToken(c.UserID, c.Roles)
	if err != nil {
		return "", "", 0, err
	}
	refresh, _, err = s.signer.IssueRefreshToken(c.UserID)
	if err != nil {
		return "", "", 0, err
	}
	return access, refresh, s.signer.AccessTTLSeconds(), nil
}

// GetUser loads a user profile by ID.
// GetUser загружает профиль пользователя по ID.
// PasswordHash is loaded but must not be exposed via gRPC (handler strips it).
// PasswordHash загружается, но не должен отдаваться через gRPC (handler его отсекает).
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
