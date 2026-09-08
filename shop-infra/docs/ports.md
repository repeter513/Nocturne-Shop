# Порты локального стека

Единая точка входа — **Envoy**: фронт и HTTP API на одном порту.

| Сервис | Порт (host) | URL |
|--------|-------------|-----|
| **shop (UI + API)** | 8080 | http://localhost:8080 |
| BFF health | 8080 | http://localhost:8080/health |
| API | 8080 | http://localhost:8080/api/v1/... |
| envoy (gRPC gateway) | 8443 | grpc через authority (см. ниже) |
| envoy admin | 9901 | http://localhost:9901 |
| PostgreSQL | 5432 | postgres://shop:shop@localhost:5432 |
| Adminer | 8089 | http://localhost:8089 |

gRPC-сервисы с хоста не проброшены — только через Envoy `:8443`:

```bash
grpcurl -plaintext -authority catalog.local localhost:8443 list
```

Порт `:8080` настраивается через `ENVOY_HTTP_PORT` в `shop-infra/.env`.

## Базы данных

| БД | Сервис |
|----|--------|
| auth_db | auth |
| catalog_db | catalog |
| cart_db | cart |
| order_db | order |
| payment_db | payment |
| notification_db | notification |
