# shop-auth

gRPC-сервис аутентификации: регистрация, логин, JWT, данные пользователя.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт: [shop-proto `auth.v1.AuthService`](../shop-proto/proto/auth/v1/auth.proto)

Локальный стек: [shop-infra](../shop-infra/README.md) (порт `8081`)

[shop-cart](../shop-cart/README.md) и [shop-order](../shop-order/README.md) проверяют access JWT (`JWT_SECRET` тот же). [shop-BFF](../shop-BFF/README.md) / [shop-web](../shop-web/README.md) ходят сюда за токеном.

## Стек

- Go 1.26
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)
- JWT (`golang-jwt`)

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL
- Репозиторий [shop-proto](../shop-proto/README.md)

Либо весь стек из [shop-infra](../shop-infra/README.md): `make up`.

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC-сервера |
| `DATABASE_URL` | да | DSN PostgreSQL |
| `JWT_SECRET` | да | Секрет JWT, **base64** |
| `JWT_ACCESS_TTL` | да | TTL access-токена (`15m`) |
| `JWT_REFRESH_TTL` | да | TTL refresh-токена (`168h`) |
| `LOG_LEVEL` | да | Уровень логирования |

В compose значения задаются в [shop-infra `env/auth.env`](../shop-infra/env/auth.env.example).

### Миграции и запуск

```bash
make migrate-up
make run
```

## gRPC API

| RPC | Описание |
|---|---|
| `RegisterUser` | Регистрация по email и паролю |
| `LoginUser` | Вход, access/refresh JWT |
| `ValidateToken` | Проверка access-токена |
| `RefreshToken` | Обновление пары токенов |
| `GetUserInfo` | Пользователь по ID |

```bash
grpcurl -plaintext localhost:8081 auth.v1.AuthService/LoginUser
```

## Docker

```bash
cd ../shop-infra && make up
```

## Make-команды

| Команда | Описание |
|---|---|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make migrate-up` / `migrate-down` | Миграции |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
