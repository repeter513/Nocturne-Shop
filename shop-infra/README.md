# shop-infra

Docker Compose для локального запуска всего магазина: инфра + микросервисы + BFF + web + **Envoy как единая точка входа**.

**Экосистема:** [infra](README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

| Компонент | Роль | Доступ с хоста |
|-----------|------|----------------|
| **Envoy** | UI + HTTP API + gRPC gateway | `:8090`, `:8443` |
| [shop-web](../shop-web/README.md) | React (статика) | через Envoy `/` |
| [shop-BFF](../shop-BFF/README.md) | HTTP JSON API | через Envoy `/api/` |
| gRPC-сервисы | auth, catalog, cart, order, payment | через Envoy `:8443` |

## Что внутри

- **Инфра:** PostgreSQL 16, Adminer
- **Сервисы:** auth, catalog, cart, order, payment (только docker-сеть)
- **BFF + web:** без портов на хост — трафик через Envoy
- **Gateway:** `envoy/envoy.yaml` — `/` → web, `/api/` → BFF, gRPC по authority на `:8443`
- **Init:** `postgres/init` создаёт БД
- **Env-шаблоны:** `.env.example` и `env/*.env.example` — фиктивные значения, чтобы было понятно, что куда подставлять.

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

```bash
make init      # .env и env/*.env из шаблонов (если ещё нет)
make up        # postgres + сервисы + bff + web + envoy (--build)
make migrate   # миграции (после up)
```

**Магазин:** http://localhost:8090  
**Health BFF:** http://localhost:8090/health  
**Adminer:** http://localhost:8089

Только инфра (без сервисов):

```bash
make infra-up
```

### Checkout flow (E2E)

1. http://localhost:8090 — регистрация / логин
2. Добавить товары в корзину
3. Оформить заказ (`CreateOrder`)
4. Оплатить или отменить

## Команды

| Команда | Описание |
|---------|----------|
| `make up` | Поднять весь стек |
| `make down` | Остановить контейнеры |
| `make infra-up` | postgres, adminer |
| `make migrate` | `scripts/migrate-all.sh` |
| `make logs` | Логи compose |
| `make init` / `make env` | Создать `.env` и `env/*.env` из шаблонов |
| `make envoy-reload` | Перезагрузить конфиг Envoy |

## Порты (.env)

| Переменная | Default | Описание |
|------------|---------|----------|
| `ENVOY_HTTP_PORT` | `8090` | UI + HTTP API |
| `ENVOY_GRPC_PORT` | `8443` | gRPC gateway (host → container :443) |
| `ENVOY_ADMIN_PORT` | `9901` | Envoy admin |
| `ADMINER_PORT` | `8089` | Adminer |

См. [docs/ports.md](docs/ports.md).

## Env для order

Файл `env/order.env` должен содержать:

```env
GRPC_PORT=8084
DATABASE_URL=postgres://shop:shop@postgres:5432/order_db?sslmode=disable
CART_GRPC_ADDR=cart:8083
CATALOG_GRPC_ADDR=catalog:8082
PAYMENT_GRPC_ADDR=payment:8086
JWT_SECRET=ZGV2LXNlY3JldC1jaGFuZ2UtbWU=
```

## Докуменция сервисов

- [shop-proto](../shop-proto/README.md) — контракты
- [shop-auth](../shop-auth/README.md)
- [shop-catalog](../shop-catalog/README.md)
- [shop-cart](../shop-cart/README.md)
- [shop-order](../shop-order/README.md)
- [shop-payment](../shop-payment/README.md)
- [shop-BFF](../shop-BFF/README.md)
- [shop-web](../shop-web/README.md)
