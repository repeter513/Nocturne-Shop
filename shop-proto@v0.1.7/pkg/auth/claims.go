// JWT claim types and audience constants for the shop microservices.
// Типы JWT-claims и константы audience для микросервисов магазина.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// Issuer is the JWT iss claim — identifies shop-auth as token issuer.
	// Issuer — claim iss JWT; идентифицирует shop-auth как издателя токена.
	Issuer = "shop-auth"

	// AudienceAuth is the audience for auth-service endpoints.
	// AudienceAuth — audience для эндпоинтов auth-сервиса.
	AudienceAuth = "shop-auth"

	// AudienceCatalog is the audience for catalog-service endpoints.
	// AudienceCatalog — audience для эндпоинтов catalog-сервиса.
	AudienceCatalog = "shop-catalog"

	// AudienceCart is the audience for cart-service endpoints.
	// AudienceCart — audience для эндпоинтов cart-сервиса.
	AudienceCart = "shop-cart"

	// AudienceOrder is the audience for order-service endpoints.
	// AudienceOrder — audience для эндпоинтов order-сервиса.
	AudienceOrder = "shop-order"

	// AudiencePayment is the audience for payment-service endpoints.
	// AudiencePayment — audience для эндпоинтов payment-сервиса.
	AudiencePayment = "shop-payment"
)

// AccessAudiences lists all services that accept access tokens.
// AccessAudiences — список сервисов, принимающих access-токены.
var AccessAudiences = []string{
	AudienceAuth,
	AudienceCatalog,
	AudienceCart,
	AudienceOrder,
	AudiencePayment,
}

// Claims extends jwt.RegisteredClaims with shop-specific fields.
// Claims расширяет jwt.RegisteredClaims полями, специфичными для магазина.
type Claims struct {
	// UserID is the authenticated user's primary key.
	// UserID — первичный ключ аутентифицированного пользователя.
	UserID int64 `json:"uid"`

	// Type is "access" or "refresh"; prevents token type confusion.
	// Type — "access" или "refresh"; предотвращает подмену типа токена.
	Type string `json:"type"`

	// Roles holds optional RBAC role names (e.g. "admin").
	// Roles — опциональные имена RBAC-ролей (например, "admin").
	Roles []string `json:"roles,omitempty"`

	jwt.RegisteredClaims
}

// newClaims builds a signed JWT payload with expiry and optional JTI.
// newClaims формирует payload JWT с expiry и опциональным JTI.
func newClaims(userID int64, typ string, ttl time.Duration, roles []string, issuer string, audience []string, withJTI bool) Claims {
	now := time.Now()
	rc := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    issuer,
		Audience:  audience,
	}
	// Refresh tokens require a unique JTI for future revocation support.
	// Refresh-токены требуют уникальный JTI для будущей поддержки отзыва.
	if withJTI {
		rc.ID = newJTI()
	}
	return Claims{
		UserID:           userID,
		Type:             typ,
		Roles:            roles,
		RegisteredClaims: rc,
	}
}

// newJTI generates a 32-char hex random token ID.
// newJTI генерирует 32-символьный hex ID токена.
func newJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
