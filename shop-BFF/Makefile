-include .env
export

.PHONY: run build tidy fmt vet test

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

test:
	go test ./...
