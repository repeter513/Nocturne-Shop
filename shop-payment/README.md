# shop-payment

gRPC-сервис платежей: создание и просмотр платежей по заказам.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto@v0.1.7/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт: [shop-proto `payment.v1.PaymentService`](../shop-proto@v0.1.7/proto/payment/v1/payment.proto) · сборка: `github.com/repeter513/shop-proto` **v0.2.7** (GitHub)

Вызывается из [shop-order](../shop-order/README.md) в `PayOrder` (`CreatePayment`).  
Локальный стек: [shop-infra](../shop-infra/README.md) (gRPC порт `8086`, env: [`env/payment.env.example`](../shop-infra/env/payment.env.example))

## Возможности

- Создать платёж по `order_id` (сумма и `user_id` — из [shop-order](../shop-order/README.md))
- Идемпотентность: один платёж на `order_id` (`UNIQUE` в БД)
- Получить платёж по ID, список с фильтром по `order_id`
- `VoidPayment` — отмена успешного платежа по `order_id` (вызывает [shop-order](../shop-order/README.md) при сбое checkout)
- MVP: `CreatePayment` всегда возвращает `SUCCESS`

Все RPC требуют metadata `authorization: Bearer <access_token>`. `user_id` для `ListPayments` — из JWT.

## Стек

- Go 1.26.3
- gRPC + protobuf ([shop-proto](../shop-proto@v0.1.7/README.md))
- PostgreSQL (`pgx`)
- gRPC-клиент к order
- Structured JSON logs (`log/slog`, `internal/logx`)

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL (`payment_db`)
- Ed25519 `public.pem`
- Запущенный [shop-order](../shop-order/README.md) на `8084`

### Конфигурация

```bash
cp .env.example .env
```

| Переменная | Обязательная | Описание |
|---|---|---|
| `GRPC_PORT` | да | Порт gRPC (`8086`) |
| `DATABASE_URL` | да | DSN PostgreSQL (`payment_db`) |
| `LOG_LEVEL` | нет | По умолчанию `info` |
| `JWT_PUBLIC_KEY_PATH` | да | PEM Ed25519 public key |
| `ORDER_GRPC_ADDR` | да | Адрес order-сервиса |

### Миграции и запуск

```bash
make migrate-up
make run
```

## gRPC API

| RPC | Описание |
|---|---|
| `CreatePayment` | Создать платёж по `order_id` |
| `GetPayment` | Платёж по ID |
| `ListPayments` | Список платежей пользователя (фильтр `order_id`) |
| `VoidPayment` | Отменить SUCCESS-платёж по `order_id` |

```bash
grpcurl -plaintext -H 'authorization: Bearer TOKEN' \
  -d '{"order_id":1}' \
  localhost:8086 payment.v1.PaymentService/CreatePayment
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
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откат |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
