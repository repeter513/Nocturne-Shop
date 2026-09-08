# shop-BFF

HTTP-прослойка (Backend for Frontend) между [shop-web](../shop-web/README.md) и gRPC-микросервисами.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](README.md) · [web](../shop-web/README.md)

Локальный стек: [shop-infra](../shop-infra/README.md) (host `:8090`, env: [`env/bff.env.example`](../shop-infra/env/bff.env.example))

## Стек

- Go 1.26
- HTTP (`net/http`)
- gRPC-клиенты к auth, catalog, cart, order, payment ([shop-proto](../shop-proto/README.md))

## Запуск локально

```bash
cp .env.example .env
go run ./cmd/server
```

Для [shop-web](../shop-web/README.md) `npm run dev` нужен `HTTP_PORT=8090` (Vite проксирует на `:8090`).

## Docker (через shop-infra)

```bash
cd ../shop-infra && make up
```

BFF с host: http://localhost:8090 (внутри контейнера `:8080`, снаружи `BFF_HTTP_PORT=8090`).  
Фронт на `:3000` проксирует `/api/` на BFF через docker-сеть.

## API

Все эндпоинты под префиксом `/api/v1`. Защищённые маршруты требуют `Authorization: Bearer <access_token>`.

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| GET | `/health` | — | healthcheck |
| POST | `/api/v1/auth/register` | — | регистрация |
| POST | `/api/v1/auth/login` | — | вход |
| POST | `/api/v1/auth/refresh` | — | обновление токена |
| GET | `/api/v1/auth/me` | ✓ | текущий пользователь |
| GET | `/api/v1/products` | — | список товаров (`page`, `pageSize`, `categoryId`) |
| GET | `/api/v1/products/{id}` | — | товар |
| GET | `/api/v1/products/{id}/stock` | — | остаток |
| GET | `/api/v1/categories` | — | категории |
| GET | `/api/v1/cart` | ✓ | корзина |
| POST | `/api/v1/cart/items` | ✓ | добавить в корзину |
| PUT | `/api/v1/cart/items/{productId}` | ✓ | изменить количество |
| DELETE | `/api/v1/cart/items/{productId}` | ✓ | удалить позицию |
| DELETE | `/api/v1/cart` | ✓ | очистить корзину |
| POST | `/api/v1/orders` | ✓ | оформить заказ (`pending` + резерв) |
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
| `AUTH_GRPC_ADDR` | адрес shop-auth |
| `CATALOG_GRPC_ADDR` | адрес shop-catalog |
| `CART_GRPC_ADDR` | адрес shop-cart |
| `ORDER_GRPC_ADDR` | адрес shop-order |
| `PAYMENT_GRPC_ADDR` | адрес shop-payment |
| `LOG_LEVEL` | уровень логирования (default `info`) |
| `CORS_ORIGINS` | allowed origin (default `*`) |
