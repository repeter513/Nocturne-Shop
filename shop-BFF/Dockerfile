# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
RUN apk add --no-cache git
COPY --from=shop-proto . /shop-proto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod edit -replace github.com/repeter513/shop-proto=/shop-proto
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
CMD ["/server"]
