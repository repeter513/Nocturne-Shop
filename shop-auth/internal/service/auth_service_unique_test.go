package service

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	if !isUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected unique violation")
	}
	if isUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Fatal("expected not unique violation")
	}
	if isUniqueViolation(errors.New("other")) {
		t.Fatal("expected false for non-pg error")
	}
}

func TestErrInvalidCredentials(t *testing.T) {
	if !errors.Is(ErrInvalidCredentials, ErrInvalidCredentials) {
		t.Fatal("expected sentinel match")
	}
	if errors.Is(pgx.ErrNoRows, ErrInvalidCredentials) {
		t.Fatal("pgx.ErrNoRows must be mapped in Login, not compared directly")
	}
	if errors.Is(errors.New("invalid credentials"), ErrInvalidCredentials) {
		t.Fatal("must return ErrInvalidCredentials, not errors.New with same text")
	}
}
