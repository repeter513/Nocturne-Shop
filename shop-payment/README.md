# shop-payment

gRPC-сервис платежей: создание и просмотр платежей по заказам.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт: [shop-proto `payment.v1.PaymentService`](../shop-proto/proto/payment/v1/payment.proto)

Вызывается из [shop-order](../shop-order/README.md) в `PayOrder` (`CreatePayment`).  
Локальный стек: [shop-infra](../shop-infra/README.md) (gRPC порт `8086`, env: [`env/payment.env.example`](../shop-infra/env/payment.env.example))

## Возможности

- Создать платёж по `order_id`, `user_id`, `amount`
- Идемпотентность: один платёж на `order_id` (`UNIQUE` в БД)
- Получить платёж по ID, список с фильтрами
- Флаг `simulate_failure` в proto для тестов отказов

## Стек

- Go 1.26
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL (`payment_db`)
- `psql` для миграций

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC (`8086`) |
| `DATABASE_URL` | да | DSN PostgreSQL (`payment_db`) |
| `LOG_LEVEL` | нет | По умолчанию `info` |

### Миграции и запуск

```bash
make migrate-up   # 00001 schema + 00002 unique(order_id)
make run
```

## gRPC API

| RPC | Описание |
|---|---|
| `CreatePayment` | Создать платёж (статус `success` или `failed`) |
| `GetPayment` | Платёж по ID |
| `ListPayments` | Список с фильтрами `user_id`, `order_id` |

```bash
grpcurl -plaintext -d '{"order_id":1,"user_id":1,"amount":99.97}' \
  localhost:8086 payment.v1.PaymentService/CreatePayment
```

## Docker

Через [shop-infra](../shop-infra/README.md):

```bash
cd ../shop-infra
make up
```

Отдельно:

```bash
docker build -t shop-payment .
docker run --env-file .env -p 8086:8086 shop-payment
```

## Make-команды

| Команда | Описание |
|---|---|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откат |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
