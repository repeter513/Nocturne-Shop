// Tests for JWT signing, verification, audience checks, and PEM key loading.
// Тесты подписи/проверки JWT, audience и загрузки PEM-ключей.
package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"slices"
	"testing"
	"time"
)

// encodePrivateKeyPEM serializes an Ed25519 private key for LoadPrivateKeyPEM tests.
// encodePrivateKeyPEM сериализует Ed25519 приватный ключ для тестов LoadPrivateKeyPEM.
func encodePrivateKeyPEM(priv ed25519.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b}), nil
}

func encodePublicKeyPEM(pub ed25519.PublicKey) ([]byte, error) {
	b, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b}), nil
}

// TestSignerVerifier covers access/refresh issuance, audience mismatch, and tampering.
// TestSignerVerifier проверяет выдачу access/refresh, несовпадение audience и подмену.
func TestSignerVerifier(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	signer := NewSigner(priv, time.Minute, time.Hour, Issuer, AccessAudiences)
	verifier := NewVerifier(pub, Issuer, AudienceCatalog)

	token, err := signer.IssueAccessToken(42, []string{"user"})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := verifier.ParseAccess(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 {
		t.Fatalf("user id: got %d want 42", claims.UserID)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "user" {
		t.Fatalf("roles: got %v", claims.Roles)
	}
	if claims.Issuer != Issuer {
		t.Fatalf("issuer: got %q want %q", claims.Issuer, Issuer)
	}
	if !slices.Contains(claims.Audience, AudienceCatalog) {
		t.Fatalf("audience: got %v want %q", claims.Audience, AudienceCatalog)
	}

	// Wrong audience must reject the token.
	// Неверный audience должен отклонить токен.
	wrongAud := NewVerifier(pub, Issuer, "shop-unknown")
	if _, err := wrongAud.ParseAccess(token); err == nil {
		t.Fatal("token must fail for wrong audience")
	}

	refresh, jti, err := signer.IssueRefreshToken(42)
	if err != nil {
		t.Fatal(err)
	}
	if jti == "" {
		t.Fatal("refresh jti must be set")
	}

	authVerifier := NewVerifier(pub, Issuer, AudienceAuth)
	refreshClaims, err := authVerifier.ParseRefresh(refresh)
	if err != nil {
		t.Fatal(err)
	}
	if refreshClaims.ID != jti {
		t.Fatalf("jti: got %q want %q", refreshClaims.ID, jti)
	}

	// Token type confusion must be rejected.
	// Подмена типа токена должна отклоняться.
	if _, err := verifier.ParseAccess(refresh); err == nil {
		t.Fatal("refresh token must not parse as access")
	}
	if _, err := verifier.ParseRefresh(refresh); err == nil {
		t.Fatal("refresh token must not parse with catalog audience")
	}

	if _, err := verifier.ParseAccess(token + "x"); err == nil {
		t.Fatal("tampered token must fail")
	}
}

// TestLoadKeyPEM verifies round-trip PEM encode/decode and signing after load.
// TestLoadKeyPEM проверяет round-trip PEM encode/decode и подпись после загрузки.
func TestLoadKeyPEM(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	privPEM, err := encodePrivateKeyPEM(priv)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM, err := encodePublicKeyPEM(pub)
	if err != nil {
		t.Fatal(err)
	}

	loadedPriv, err := LoadPrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	loadedPub, err := LoadPublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	token, err := NewSigner(loadedPriv, time.Minute, time.Hour, Issuer, AccessAudiences).IssueAccessToken(1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewVerifier(loadedPub, Issuer, AudienceAuth).ParseAccess(token); err != nil {
		t.Fatal(err)
	}
}
