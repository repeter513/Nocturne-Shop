# shop-auth

gRPC-сервис аутентификации: регистрация, логин, JWT (Ed25519), данные пользователя.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](README.md) · [catalog](../shop-catolog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт: [shop-proto `auth.v1.AuthService`](../shop-proto/proto/auth/v1/auth.proto) (модуль `v0.2.6`)

Локальный стек: [shop-infra](../shop-infra/README.md) (порт `8081`)

[shop-cart](../shop-cart/README.md) и [shop-order](../shop-order/README.md) проверяют access JWT тем же `public.pem`. [shop-BFF](../shop-BFF/README.md) / [shop-web](../shop-web/README.md) ходят сюда за токеном.

## Стек

- Go 1.26.3
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)
- JWT Ed25519 ([shop-proto `pkg/auth`](../shop-proto/README.md#auth-pkgauth))

## JWT-ключи (Ed25519)

Один раз перед первым запуском:

```bash
mkdir -p ../.shop-keys
openssl genpkey -algorithm ED25519 -out ../.shop-keys/private.pem
openssl pkey -in ../.shop-keys/private.pem -pubout -out ../.shop-keys/public.pem
chmod 600 ../.shop-keys/private.pem
```

| Файл | Где нужен |
|---|---|
| `private.pem` | только [shop-auth](README.md) (`JWT_PRIVATE_KEY_PATH`) |
| `public.pem` | auth + cart, catalog, order, payment (`JWT_PUBLIC_KEY_PATH`) |

Пути по умолчанию — в `.env.example`. В compose каталог монтируется через `JWT_KEYS_DIR` ([shop-infra](../shop-infra/README.md)).

Ключи не коммитятся. После смены ключей — повторный login (старые токены недействительны).

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL
- Ed25519 ключи (см. выше)

Либо весь стек из [shop-infra](../shop-infra/README.md): `make init && make up`.

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC-сервера |
| `DATABASE_URL` | да | DSN PostgreSQL |
| `JWT_PRIVATE_KEY_PATH` | да | PEM Ed25519 private key (только auth) |
| `JWT_PUBLIC_KEY_PATH` | да | PEM Ed25519 public key |
| `JWT_ACCESS_TTL` | да | TTL access-токена (`15m`) |
| `JWT_REFRESH_TTL` | да | TTL refresh-токена (`168h`) |
| `LOG_LEVEL` | да | Уровень логирования |

В compose ключи монтируются из `JWT_KEYS_DIR` → `/run/jwt/` ([shop-infra](../shop-infra/README.md)).

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
grpcurl -plaintext -d '{"email":"u@example.com","password":"secret"}' \
  localhost:8081 auth.v1.AuthService/LoginUser
```

## Docker

```bash
cd ../shop-infra && make init && make up
```

## Make-команды

| Команда | Описание |
|---|---|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make migrate-up` / `migrate-down` | Миграции |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
