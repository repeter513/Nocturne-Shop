// JWT token verification and parsing with Ed25519 public key.
// Проверка и разбор JWT-токенов с Ed25519 публичным ключом.
package auth

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Verifier validates JWT signatures and standard claims.
// Verifier проверяет подписи JWT и стандартные claims.
type Verifier struct {
	// PublicKey verifies EdDSA signatures.
	// PublicKey проверяет подписи EdDSA.
	PublicKey ed25519.PublicKey

	// ExpectedIssuer must match the iss claim.
	// ExpectedIssuer должен совпадать с claim iss.
	ExpectedIssuer string

	// ExpectedAudience must appear in the aud claim list.
	// ExpectedAudience должен присутствовать в списке aud.
	ExpectedAudience string
}

// NewVerifier creates a Verifier bound to issuer and audience.
// NewVerifier создаёт Verifier, привязанный к issuer и audience.
func NewVerifier(publicKey ed25519.PublicKey, issuer, audience string) *Verifier {
	return &Verifier{
		PublicKey:        publicKey,
		ExpectedIssuer:   issuer,
		ExpectedAudience: audience,
	}
}

// Parse validates signature, issuer, audience and returns claims.
// Parse проверяет подпись, issuer, audience и возвращает claims.
func (v *Verifier) Parse(token string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithIssuer(v.ExpectedIssuer), jwt.WithAudience(v.ExpectedAudience))
	t, err := parser.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		// Reject tokens signed with any algorithm other than EdDSA.
		// Отклоняем токены, подписанные алгоритмом, отличным от EdDSA.
		if t.Method != jwt.SigningMethodEdDSA {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		return v.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseAccess accepts only tokens with type "access".
// ParseAccess принимает только токены с type "access".
func (v *Verifier) ParseAccess(token string) (*Claims, error) {
	c, err := v.Parse(token)
	if err != nil {
		return nil, err
	}
	if c.Type != "access" {
		return nil, errors.New("not an access token")
	}
	return c, nil
}

// ParseRefresh accepts only refresh tokens that include a JTI.
// ParseRefresh принимает только refresh-токены с JTI.
func (v *Verifier) ParseRefresh(token string) (*Claims, error) {
	c, err := v.Parse(token)
	if err != nil {
		return nil, err
	}
	if c.Type != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	if c.ID == "" {
		return nil, errors.New("missing jti")
	}
	return c, nil
}

// ParseAccessUserID is a convenience wrapper that extracts user ID from an access token.
// ParseAccessUserID — обёртка для извлечения user ID из access-токена.
func ParseAccessUserID(token string, v *Verifier) (int64, error) {
	c, err := v.ParseAccess(token)
	if err != nil {
		return 0, err
	}
	return c.UserID, nil
}
