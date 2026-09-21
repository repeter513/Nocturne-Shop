# shop-infra

Docker Compose для локального запуска всего магазина: инфра + микросервисы + BFF + web + **Envoy как единая точка входа**.

**Экосистема:** [infra](README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catolog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

| Компонент | Роль | Доступ с хоста |
|-----------|------|----------------|
| **Envoy** | UI + HTTP API + gRPC gateway | `:8090`, `:8443` |
| [shop-web](../shop-web/README.md) | React (статика) | через Envoy `/` |
| [shop-BFF](../shop-BFF/README.md) | HTTP JSON API | через Envoy `/api/` |
| gRPC-сервисы | auth, catalog, cart, order, payment | через Envoy `:8443` |
| **Adminer** | UI для PostgreSQL | `:8089` |
| **Postgres** | БД всех сервисов | `:5432` |

## Структура repos (multi-repo)

```
../shop-infra
../shop-proto
../shop-auth
../shop-catolog      # shop-catalog
../shop-cart
../shop-order
../shop-payment
../shop-BFF
../shop-web
```

Монорепо: [Nocturne](https://github.com/repeter513/Nocturne)

## JWT (Ed25519)

Перед первым `make up` создай ключи (если ещё нет):

```bash
mkdir -p ../.shop-keys
openssl genpkey -algorithm ED25519 -out ../.shop-keys/private.pem
openssl pkey -in ../.shop-keys/private.pem -pubout -out ../.shop-keys/public.pem
chmod 600 ../.shop-keys/private.pem
```

В `.env` укажи `JWT_KEYS_DIR=../.shop-keys` (см. `.env.example`) — путь относительно `docker-compose.yml`. Compose монтирует каталог в контейнеры как `/run/jwt/`; в `env/*.env` пути внутри контейнера: `/run/jwt/*.pem`. `~` в compose не раскрывается.

Подробнее: [shop-auth](../shop-auth/README.md#jwt-ключи-ed25519).

## Быстрый старт

Перед `make up` нужны Ed25519 ключи (см. выше) и `.env` в sibling-репозиториях (auth, cart, catalog, order, payment, BFF).

```bash
make init    # .env здесь + env/*.env из example
make up      # build + start
make migrate # миграции всех сервисов
```

**Магазин:** http://localhost:8090  
**Health BFF:** http://localhost:8090/health  
**Adminer:** http://localhost:8089

## Make-команды

| Команда | Описание |
|---------|----------|
| `make init` | Создать `.env` и `env/*.env` из example |
| `make up` | Собрать и поднять весь стек |
| `make down` | Остановить контейнеры |
| `make infra-up` | Только postgres + adminer |
| `make migrate` | Миграции auth, catalog, cart, order, payment |
| `make logs` | `docker compose logs -f` |
| `make seed` | Seed через `../shop-seed` (если есть) |
| `make env` | Создать `.env` / `env/*.env` если отсутствуют |
| `make envoy-reload` | Перезагрузить конфиг Envoy |

## Переменные (.env)

| Переменная | Default | Описание |
|------------|---------|----------|
| `JWT_KEYS_DIR` | — | Каталог с `private.pem` / `public.pem` |
| `POSTGRES_USER` | `shop` | Пользователь PostgreSQL |
| `POSTGRES_PASSWORD` | `shop` | Пароль PostgreSQL |
| `POSTGRES_HOST` | `postgres` | Хост внутри compose |
| `POSTGRES_PORT` | `5432` | Порт PostgreSQL |
| `ADMINER_PORT` | `8089` | Adminer на хосте |
| `ENVOY_HTTP_PORT` | `8090` | UI + HTTP API |
| `ENVOY_GRPC_PORT` | `8443` | gRPC gateway |
| `ENVOY_ADMIN_PORT` | `9901` | Envoy admin |

См. [docs/ports.md](docs/ports.md).
