package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewJWT(secret []byte, accessTTL, refreshTTL time.Duration) *JWT {
	return &JWT{
		Secret:     secret,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

type Claims struct {
	UserID int64  `json:"uid"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func (j *JWT) IssueAccessToken(userID int64) (string, error) {
	return j.issue(userID, "access", j.AccessTTL)
}

func (j *JWT) IssueRefreshToken(userID int64) (string, error) {
	return j.issue(userID, "refresh", j.RefreshTTL)
}

func (j *JWT) issue(userID int64, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.Secret)
}

func (j *JWT) Parse(token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		return j.Secret, nil
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

func (j *JWT) ParseAccess(token string) (*Claims, error) {
	c, err := j.Parse(token)
	if err != nil {
		return nil, err
	}
	if c.Type != "access" {
		return nil, errors.New("not an access token")
	}
	return c, nil
}

func (j *JWT) ParseRefresh(token string) (*Claims, error) {
	c, err := j.Parse(token)
	if err != nil {
		return nil, err
	}
	if c.Type != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	return c, nil
}

func (j *JWT) AccessTTLSeconds() int64 {
	return int64(j.AccessTTL.Seconds())
}

func ParseAccessUserID(token string, secret []byte) (int64, error) {
	c, err := NewJWT(secret, 0, 0).ParseAccess(token)
	if err != nil {
		return 0, err
	}
	return c.UserID, nil
}
