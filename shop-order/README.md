# shop-order

gRPC-сервис заказов: оформление из корзины, резерв стока, оплата / отмена.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт: [shop-proto `order.v1.OrderService`](../shop-proto/proto/order/v1/order.proto)

Локальный стек: [shop-infra](../shop-infra/README.md) (gRPC порт `8084`, env: [`env/order.env.example`](../shop-infra/env/order.env.example))

Все RPC требуют metadata `authorization: Bearer <access_token>` (`JWT_SECRET` тот же, что в [shop-auth](../shop-auth/README.md)). `user_id` берётся из JWT.

## Зависимости

| Сервис | Зачем |
|--------|--------|
| [shop-cart](../shop-cart/README.md) | `GetCart`, `ClearCart` |
| [shop-catalog](../shop-catalog/README.md) | `ReserveStock`, `ConfirmReservation`, `ReleaseStock` |
| [shop-payment](../shop-payment/README.md) | `CreatePayment` |

## Checkout flow

```
CreateOrder:
  GetCart → INSERT order (pending) → ReserveStock → return pending

PayOrder (только pending, свой заказ):
  CreatePayment → ConfirmReservation → UPDATE paid + payment_id → ClearCart (best-effort)

CancelOrder (только pending, свой заказ):
  ReleaseStock → UPDATE cancelled
```

При ошибке оплаты: `ReleaseStock` + статус `failed` (с `payment_id` если платёж уже создан).

## Стек

- Go 1.26
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)
- JWT (`golang-jwt`)

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL (`order_db`)
- Запущенные [cart](../shop-cart/README.md) (`8083`), [catalog](../shop-catalog/README.md) (`8082`), [payment](../shop-payment/README.md) (`8086`)

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC (`8084`) |
| `DATABASE_URL` | да | DSN PostgreSQL (`order_db`) |
| `LOG_LEVEL` | нет | По умолчанию `info` |
| `CART_GRPC_ADDR` | да | Адрес cart |
| `CATALOG_GRPC_ADDR` | да | Адрес catalog |
| `PAYMENT_GRPC_ADDR` | да | Адрес payment |
| `JWT_SECRET` | да | base64 secret, тот же что в [shop-auth](../shop-auth/README.md) |

### Миграции и запуск

```bash
make migrate-up
make run
```

## gRPC API

| RPC | Описание |
|---|---|
| `CreateOrder` | Заказ из корзины: резерв стока, статус `pending` |
| `PayOrder` | Оплатить pending-заказ |
| `CancelOrder` | Отменить pending-заказ, снять резерв |
| `GetOrder` | Заказ по ID |
| `ListOrders` | Список заказов пользователя |

```bash
grpcurl -plaintext -H 'authorization: Bearer TOKEN' \
  localhost:8084 order.v1.OrderService/CreateOrder
grpcurl -plaintext -H 'authorization: Bearer TOKEN' \
  -d '{"order_id":1}' localhost:8084 order.v1.OrderService/PayOrder
grpcurl -plaintext -H 'authorization: Bearer TOKEN' \
  -d '{"order_id":1}' localhost:8084 order.v1.OrderService/GetOrder
```

## Docker

Через [shop-infra](../shop-infra/README.md):

```bash
cd ../shop-infra
make up
```

## Make-команды

| Команда | Описание |
|---|---|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make migrate-up` | Схема + seed |
| `make migrate-down` | Откат |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
