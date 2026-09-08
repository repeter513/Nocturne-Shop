-include .env
export

.PHONY: run build tidy fmt vet migrate-up migrate-down migrate-seed

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
	psql "$(DATABASE_URL)" -f migrations/00001_cart.up.sql
	psql "$(DATABASE_URL)" -f migrations/00002_seed.up.sql

migrate-down:
	psql "$(DATABASE_URL)" -f migrations/00002_seed.down.sql
	psql "$(DATABASE_URL)" -f migrations/00001_cart.down.sql

migrate-seed:
	psql "$(DATABASE_URL)" -f migrations/00002_seed.up.sql
