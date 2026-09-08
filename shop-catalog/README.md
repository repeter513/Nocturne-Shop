# shop-catalog

gRPC-сервис каталога магазина: товары, категории, остатки и резервирование стока под заказ.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](../shop-web/README.md)

Контракт API: [shop-proto `catalog.v1.CatalogService`](../shop-proto/proto/catalog/v1/catalog.proto)

Используется [shop-cart](../shop-cart/README.md) (цены, сток) и [shop-order](../shop-order/README.md) (`CreateOrder` → резерв, `PayOrder` → confirm, `CancelOrder` → release).

## Возможности

- Получение товара и списка товаров (с пагинацией и фильтром по категории)
- Список категорий
- Доступный остаток с учётом активных резервов
- Резервирование, снятие и подтверждение резерва по `order_id`
- Фоновая очистка просроченных резервов

## Стек

- Go 1.26 
- gRPC + protobuf ([shop-proto](../shop-proto/README.md))
- PostgreSQL (`pgx`)

## Быстрый старт

```bash
cp .env.example .env
make migrate-up
make run
```

Весь стек: [shop-infra](../shop-infra/README.md) → `make up` (gRPC порт `8082`).

### Конфигурация

| Переменная | Обязательная | По умолчанию | Описание |
|---|---|---|---|
| `GRPC_PORT` | да | — | Порт gRPC |
| `DATABASE_URL` | да | — | DSN PostgreSQL (`catalog_db`) |
| `LOG_LEVEL` | нет | `info` | Уровень логирования |
| `RESERVATION_TTL` | нет | `5m` | TTL резерва |
| `CLEANUP_INTERVAL` | нет | `1m` | Интервал очистки просроченных резервов |

## gRPC API

| RPC | Описание |
|---|---|
| `GetProduct` | Товар по ID |
| `ListProducts` | Список активных товаров |
| `ListCategories` | Все категории |
| `GetStock` | Доступный остаток |
| `ReserveStock` | Зарезервировать товар под `order_id` |
| `ReleaseStock` | Снять резерв |
| `ConfirmReservation` | Подтвердить — списать сток |

```bash
grpcurl -plaintext localhost:8082 catalog.v1.CatalogService/ListProducts
grpcurl -plaintext -d '{"product_id":1,"quantity":2,"order_id":1001}' \
  localhost:8082 catalog.v1.CatalogService/ReserveStock
```

## Модель стока

- `products.stock` — физический остаток
- Активный резерв **не уменьшает** `stock`, но уменьшает **доступный** остаток
- При `ConfirmReservation` сток списывается

```
ReserveStock → active → ConfirmReservation → confirmed (stock -= qty)
                      → ReleaseStock       → released
                      → TTL истёк          → expired
```

**Один `order_id` — одна резервация.** Повторные вызовы `ReserveStock` для того же `order_id` **мерджат** позиции (нужно для multi-item checkout из [shop-order](../shop-order/README.md)).

## Docker

```bash
cd ../shop-infra && make up
```

## Make-команды

| Команда | Описание |
|---|---|
| `make run` / `make build` | Запуск / сборка |
| `make migrate-up` | Схема + seed |
| `make migrate-down` / `migrate-seed` | Откат / seed |
| `make fmt` / `make vet` / `make tidy` | Форматирование и проверки |
