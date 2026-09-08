# shop-infra

Docker Compose для локального запуска всего магазина: инфра + микросервисы + BFF + web.

**Экосистема:** [infra](README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

| Репо | Роль | Порт |
|------|------|------|
| [shop-proto](../shop-proto/README.md) | gRPC-контракты | — |
| [shop-auth](../shop-auth/README.md) | auth | `8081` |
| [shop-catalog](../shop-catalog/README.md) | catalog | `8082` |
| [shop-cart](../shop-cart/README.md) | cart | `8083` |
| [shop-order](../shop-order/README.md) | order | `8084` |
| [shop-payment](../shop-payment/README.md) | payment | `8086` |
| [shop-BFF](../shop-BFF/README.md) | HTTP API для фронта | `8090` (host) |
| [shop-web](../shop-web/README.md) | фронт (nginx) | `3000` |

## Что внутри

- **Инфра:** PostgreSQL 16, Adminer
- **Сервисы:** auth, catalog, cart, order, payment (build из соседних папок монорепо)
- **BFF + web:** HTTP API на host `:8090` (внутри контейнера `:8080`), UI на `:3000`
- **Gateway:** Envoy — единая точка входа gRPC на `:443` (конфиг: `envoy/envoy.yaml`)
- **Init:** `postgres/init` создаёт БД
- **Env-шаблоны:** `.env.example` и `env/*.env.example`
## Структура repos

**Монорепо [Nocturne](../README.md)** — все сервисы уже в одном дереве:

```
Nocturne/
  shop-infra
  shop-proto
  shop-auth
  shop-catalog
  shop-cart
  shop-order
  shop-payment
  shop-BFF
  shop-web
```

Запуск из корня монорепо: `make up` (делегирует сюда).

## Быстрый старт

Из корня [Nocturne](../README.md) или из `shop-infra`:

```bash
make init      # .env и env/*.env из шаблонов (если ещё нет)
make up        # postgres + все сервисы + bff + web (--build)
make migrate   # миграции всех сервисов (после up, когда postgres готов)
```

`make up` и `make infra-up` сами вызывают `make env`, если файлов ещё нет.

UI: http://localhost:3000  
BFF: http://localhost:8090/health

Только инфра (без сервисов):

```bash
make infra-up
```

### Checkout flow (E2E)

1. [shop-web](../shop-web/README.md) / [shop-BFF](../shop-BFF/README.md) — регистрация / логин
2. [shop-cart](../shop-cart/README.md) — добавить товары
3. [shop-order](../shop-order/README.md) — `CreateOrder` (корзина → резерв стока, статус `pending`)
4. [shop-order](../shop-order/README.md) — `PayOrder` (payment → confirm стока) или `CancelOrder`

## Команды

| Команда | Описание |
|---------|----------|
| `make up` | Поднять infra + все сервисы + bff + web |
| `make down` | Остановить и удалить контейнеры |
| `make infra-up` | postgres, adminer |
| `make migrate` | `scripts/migrate-all.sh` — миграции auth, catalog, cart, order, payment |
| `make logs` | Логи всех контейнеров |
| `make init` / `make env` | Создать `.env` и `env/*.env` из шаблонов |

## Порты и БД

См. [docs/ports.md](docs/ports.md).

Дефолты Postgres: `shop` / `shop`. Сервисы — gRPC; BFF — HTTP.

## Env для order

Файл `env/order.env` должен содержать:

```env
GRPC_PORT=8084
DATABASE_URL=postgres://shop:shop@postgres:5432/order_db?sslmode=disable
CART_GRPC_ADDR=cart:8083
CATALOG_GRPC_ADDR=catalog:8082
PAYMENT_GRPC_ADDR=payment:8086
JWT_SECRET=<тот же base64, что в auth>
```

## Документация сервисов

- [shop-proto](../shop-proto/README.md) — контракты
- [shop-auth](../shop-auth/README.md)
- [shop-catalog](../shop-catalog/README.md)
- [shop-cart](../shop-cart/README.md)
- [shop-order](../shop-order/README.md)
- [shop-payment](../shop-payment/README.md)
- [shop-BFF](../shop-BFF/README.md)
- [shop-web](../shop-web/README.md)
