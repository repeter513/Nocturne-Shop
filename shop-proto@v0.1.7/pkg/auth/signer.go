// JWT token signing with Ed25519 for access and refresh tokens.
// Подпись JWT-токенов Ed25519 для access и refresh.
package auth

import (
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Signer holds the private key and TTL settings for token issuance.
// Signer хранит приватный ключ и настройки TTL для выдачи токенов.
type Signer struct {
	// PrivateKey signs tokens with EdDSA algorithm.
	// PrivateKey подписывает токены алгоритмом EdDSA.
	PrivateKey ed25519.PrivateKey

	// AccessTTL is how long access tokens remain valid.
	// AccessTTL — время жизни access-токенов.
	AccessTTL time.Duration

	// RefreshTTL is how long refresh tokens remain valid.
	// RefreshTTL — время жизни refresh-токенов.
	RefreshTTL time.Duration

	// Issuer is written into the iss claim.
	// Issuer записывается в claim iss.
	Issuer string

	// Audience is the list of aud claims for access tokens.
	// Audience — список aud claims для access-токенов.
	Audience []string
}

// NewSigner creates a Signer with the given key material and TTLs.
// NewSigner создаёт Signer с заданными ключами и TTL.
func NewSigner(privateKey ed25519.PrivateKey, accessTTL, refreshTTL time.Duration, issuer string, audience []string) *Signer {
	return &Signer{
		PrivateKey: privateKey,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
		Issuer:     issuer,
		Audience:   audience,
	}
}

// IssueAccessToken creates a short-lived access token for API calls.
// IssueAccessToken создаёт короткоживущий access-токен для API-вызовов.
func (s *Signer) IssueAccessToken(userID int64, roles []string) (string, error) {
	return s.issue(userID, "access", s.AccessTTL, roles, s.Audience, false)
}

// IssueRefreshToken creates a long-lived refresh token with a unique JTI.
// IssueRefreshToken создаёт долгоживущий refresh-токен с уникальным JTI.
func (s *Signer) IssueRefreshToken(userID int64) (token, jti string, err error) {
	// Refresh audience is limited to the issuer (auth service only).
	// Audience refresh ограничен issuer (только auth-сервис).
	claims := newClaims(userID, "refresh", s.RefreshTTL, nil, s.Issuer, []string{s.Issuer}, true)
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token, err = t.SignedString(s.PrivateKey)
	if err != nil {
		return "", "", err
	}
	return token, claims.ID, nil
}

// issue is the internal helper that signs a token of the given type.
// issue — внутренний хелпер подписи токена заданного типа.
func (s *Signer) issue(userID int64, typ string, ttl time.Duration, roles []string, audience []string, withJTI bool) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, newClaims(userID, typ, ttl, roles, s.Issuer, audience, withJTI))
	return t.SignedString(s.PrivateKey)
}

// AccessTTLSeconds returns access token lifetime in seconds for API responses.
// AccessTTLSeconds возвращает время жизни access-токена в секундах для API-ответов.
func (s *Signer) AccessTTLSeconds() int64 {
	return int64(s.AccessTTL.Seconds())
}
