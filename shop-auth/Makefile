-include .env
export

.PHONY: run build tidy fmt vet migrate-up migrate-down

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

migrate-up:
	psql "$(DATABASE_URL)" -f migrations/00001_users.up.sql

migrate-down:
	psql "$(DATABASE_URL)" -f migrations/00001_users.down.sql
