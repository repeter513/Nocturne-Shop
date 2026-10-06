# shop-BFF

HTTP-прослойка (Backend for Frontend) между [shop-web](../shop-web/README.md) и gRPC-микросервисами.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto@v0.1.7/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](README.md) · [web](../shop-web/README.md)

В compose **нет порта на хост** — HTTP API доступен через [Envoy](../shop-infra/envoy/envoy.yaml):

- http://localhost:8090/health
- http://localhost:8090/api/v1/...

## Стек

- Go 1.26.3
- HTTP (`net/http`)
- gRPC-клиенты к auth, catalog, cart, order, payment — модуль [shop-proto](https://github.com/repeter513/shop-proto) **v0.2.7** (GitHub); `.proto` в [../shop-proto@v0.1.7](../shop-proto@v0.1.7/README.md)
- Structured JSON logs (`log/slog`, `internal/logx`)
- JWT access проверяется локально по `public.pem` (без `ValidateToken` RPC)
- Rate limit по IP: register/login (`AUTH_RATE_LIMIT_PER_MIN`), отдельно refresh (`AUTH_REFRESH_RATE_LIMIT_PER_MIN`); на Envoy — отдельный bucket для `POST /api/v1/auth/refresh`

## Запуск локально (без Docker)

```bash
cp .env.example .env
go run ./cmd/server
```

`.env.example` пустой — для compose используй [`shop-infra/env/bff.env`](../shop-infra/env/bff.env.example).

Слушает `:8080` внутри контейнера (`HTTP_PORT`). С хоста — через Envoy `:8090`.

## Docker (через shop-infra)

```bash
cd ../shop-infra && make up
```

BFF доступен только внутри docker-сети; снаружи — через Envoy `/api/` и `/health`.

## API

Все эндпоинты под префиксом `/api/v1`. Защищённые маршруты требуют `Authorization: Bearer <access_token>`.

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/health` | — | healthcheck |
| POST | `/api/v1/auth/register` | — | регистрация |
| POST | `/api/v1/auth/login` | — | вход |
| POST | `/api/v1/auth/refresh` | — | новая пара JWT (stateless; см. [shop-auth](../shop-auth/README.md#jwt-stateless-only)) |
| GET | `/api/v1/auth/me` | ✓ | текущий пользователь |
| GET | `/api/v1/products` | — | список товаров |
| GET | `/api/v1/products/{id}` | — | товар |
| GET | `/api/v1/products/{id}/stock` | — | остаток |
| GET | `/api/v1/categories` | — | категории |
| GET | `/api/v1/cart` | ✓ | корзина |
| POST | `/api/v1/cart/items` | ✓ | добавить в корзину |
| PUT | `/api/v1/cart/items/{productId}` | ✓ | изменить количество |
| DELETE | `/api/v1/cart/items/{productId}` | ✓ | удалить позицию |
| DELETE | `/api/v1/cart` | ✓ | очистить корзину |
| POST | `/api/v1/orders` | ✓ | оформить заказ |
| POST | `/api/v1/orders/{id}/pay` | ✓ | оплатить заказ |
| POST | `/api/v1/orders/{id}/cancel` | ✓ | отменить заказ |
| GET | `/api/v1/orders` | ✓ | список заказов |
| GET | `/api/v1/orders/{id}` | ✓ | заказ |
| GET | `/api/v1/payments` | ✓ | список платежей |
| GET | `/api/v1/payments/{id}` | ✓ | платёж |

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `HTTP_PORT` | HTTP-порт (default `8080`) |
| `AUTH_GRPC_ADDR` | адрес shop-auth (`auth:8081` в compose) |
| `CATALOG_GRPC_ADDR` | адрес shop-catalog (`catalog:8082`) |
| `CART_GRPC_ADDR` | адрес shop-cart (`cart:8083`) |
| `ORDER_GRPC_ADDR` | адрес shop-order (`order:8084`) |
| `PAYMENT_GRPC_ADDR` | адрес shop-payment (`payment:8086`) |
| `LOG_LEVEL` | уровень логирования (default `info`) |
| `CORS_ORIGINS` | allowed origin (default `*`) |
| `JWT_PUBLIC_KEY_PATH` | PEM Ed25519 public key (как у cart/order) |
| `AUTH_RATE_LIMIT_PER_MIN` | лимит register/login на IP в минуту (default `30`) |
| `AUTH_REFRESH_RATE_LIMIT_PER_MIN` | лимит refresh на IP в минуту (default `10`) |

Шаблон для compose: [`env/bff.env.example`](../shop-infra/env/bff.env.example)

## Make-команды

| Команда | Описание |
|---------|----------|
| `make run` | Запуск сервера |
| `make build` | Сборка в `bin/server` |
| `make test` | `go test ./...` |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
