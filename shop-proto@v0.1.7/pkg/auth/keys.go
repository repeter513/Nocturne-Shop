// Loading Ed25519 keys from PEM-encoded files.
// Загрузка Ed25519 ключей из PEM-файлов.
package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// LoadPrivateKeyPEM parses a PKCS#8 PEM block into an Ed25519 private key.
// LoadPrivateKeyPEM разбирает PEM-блок PKCS#8 в Ed25519 приватный ключ.
func LoadPrivateKeyPEM(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid pem")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pk, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not ed25519 private key")
	}
	return pk, nil
}

// LoadPublicKeyPEM parses a PKIX PEM block into an Ed25519 public key.
// LoadPublicKeyPEM разбирает PEM-блок PKIX в Ed25519 публичный ключ.
func LoadPublicKeyPEM(data []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid pem")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pk, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not ed25519 public key")
	}
	return pk, nil
}
