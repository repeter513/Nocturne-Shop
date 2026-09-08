# shop-cart

gRPC-сервис корзины: добавление, изменение, удаление позиций и просмотр корзины пользователя.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт API: [shop-proto `cart.v1.CartService`](../shop-proto/proto/cart/v1/cart.proto)

Локальный стек: [shop-infra](../shop-infra/README.md) (gRPC порт `8083`)

Корзина: [shop-order](../shop-order/README.md) читает её в `CreateOrder` (`GetCart`) и чистит в `PayOrder` (`ClearCart`).

## Возможности

- Добавить товар в корзину (`AddToCart`)
- Изменить количество (`UpdateCartItem`, `quantity=0` удаляет позицию)
- Удалить позицию (`RemoveFromCart`)
- Получить корзину с name, price, total (`GetCart`)
- Очистить корзину (`ClearCart`)
- Проверка стока и обогащение ответа через [shop-catalog](../shop-catalog/README.md)

## Стек

- Go 1.26
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)
- gRPC-клиент к catalog

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL (`cart_db`)
- Запущенный [shop-catalog](../shop-catalog/README.md) на `8082`

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC (`8083`) |
| `DATABASE_URL` | да | DSN PostgreSQL (`cart_db`) |
| `LOG_LEVEL` | нет | По умолчанию `info` |
| `CATALOG_GRPC_ADDR` | да | Адрес catalog |
| `JWT_SECRET` | да | base64 secret, тот же что в [shop-auth](../shop-auth/README.md) |

Все RPC требуют metadata `authorization: Bearer <access_token>`.

### Миграции и запуск

```bash
make migrate-up
make run
```

## gRPC API

| RPC | Описание |
|---|---|
| `GetCart` | Корзина пользователя |
| `AddToCart` | Добавить товар (суммирует qty) |
| `UpdateCartItem` | Установить qty |
| `RemoveFromCart` | Удалить позицию |
| `ClearCart` | Очистить корзину |

```bash
grpcurl -plaintext -H 'authorization: Bearer TOKEN' -d '{}' \
  localhost:8083 cart.v1.CartService/GetCart
grpcurl -plaintext -H 'authorization: Bearer TOKEN' \
  -d '{"product_id":1,"quantity":2}' \
  localhost:8083 cart.v1.CartService/AddToCart
```

`user_id` берётся из JWT, не из body.

## Docker

```bash
cd ../shop-infra && make up
```

Миграции в образ не входят — применяй отдельно (`make migrate-up`).

## Make-команды

| Команда | Описание |
|---|---|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make migrate-up` | Схема + seed |
| `make migrate-down` / `migrate-seed` | Откат / seed |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
