# Nocturne Shop — монорепо (демо)

Монорепозиторий интернет-магазина **Nocturne**: gRPC-микросервисы на Go, BFF, React-фронт, PostgreSQL.  
Собран для демонстрации — чтобы любой мог быстро поднять весь стек локально.

## Структура

```
Nocturne/
├── shop-proto/      gRPC-контракты (protobuf)
├── shop-infra/      Docker Compose, Postgres, Envoy, миграции
├── shop-auth/       аутентификация, JWT
├── shop-catalog/    каталог товаров
├── shop-cart/       корзина
├── shop-order/      заказы
├── shop-payment/    оплата
├── shop-BFF/        HTTP API для фронта
└── shop-web/        React + Vite UI
```

## Быстрый старт

### 1. Требования

- Docker и Docker Compose v2

### 2. Запуск (из корня монорепо)

```bash
make init      # .env и env/*.env из шаблонов (первый раз)
make up        # postgres + сервисы + bff + web + envoy
make migrate   # миграции БД (после up, когда postgres готов)
```

В `.env.example` и `env/*.env.example` — фиктивные значения, чтобы было понятно, что куда подставлять.

Первый `make up` может занять несколько минут — идёт сборка образов.

Остановка:

```bash
make down
```

### 3. Проверка, что всё работает

**Один адрес** — UI и API через Envoy:

| Что проверить | URL | Ожидаемый результат |
|---------------|-----|---------------------|
| UI магазина | http://localhost:8090 | Страница каталога Nocturne |
| BFF health | http://localhost:8090/health | `{"status":"ok"}` |
| Каталог (API) | http://localhost:8090/api/v1/products | JSON со списком товаров |
| Adminer (БД) | http://localhost:8089 | Вход: `shop` / `shop` |

```bash
curl http://localhost:8090/health
curl http://localhost:8090/api/v1/products
```



### 4. Попробовать сценарий

1. Открыть http://localhost:8090
2. Зарегистрироваться / войти
3. Добавить товар в корзину
4. Оформить заказ → оплатить или отменить

Подробнее: [shop-infra/README.md](shop-infra/README.md)

## Требования (локальная разработка)

- Go 1.26+ — только если запускаете сервисы вне Docker

## Команды (корень)

| Команда | Описание |
|---------|----------|
| `make init` | Создать `.env` и `env/*.env` из шаблонов |
| `make up` | Поднять весь стек |
| `make down` | Остановить контейнеры |
| `make migrate` | Миграции всех сервисов |
| `make logs` | Логи compose |
| `make infra-up` | Только Postgres + Adminer |

## Порты (host)

| Сервис | Порт |
|--------|------|
| **Envoy (UI + API)** | 8090 |
| Envoy gRPC gateway | 8443 |
| Adminer | 8089 |
| PostgreSQL | 5432 |

Подробнее: [shop-infra/docs/ports.md](shop-infra/docs/ports.md).

## Архитектура

```mermaid
flowchart LR
  Browser --> Envoy[Envoy :8090]
  Envoy -->|"/"| Web[shop-web]
  Envoy -->|"/api/"| BFF[shop-BFF]
  BFF --> Auth[shop-auth]
  BFF --> Catalog[shop-catalog]
  BFF --> Cart[shop-cart]
  BFF --> Order[shop-order]
  BFF --> Payment[shop-payment]
  Envoy -->|gRPC :8443| Auth
  Envoy --> Catalog
  Envoy --> Cart
  Envoy --> Order
  Envoy --> Payment
  Auth --> PG[(PostgreSQL)]
  Catalog --> PG
  Cart --> PG
  Order --> PG
  Payment --> PG
```

## Документация сервисов

- [shop-infra](shop-infra/README.md) — compose, env, миграции
- [shop-proto](shop-proto/README.md) — protobuf-контракты
- [shop-auth](shop-auth/README.md)
- [shop-catalog](shop-catalog/README.md) — каталог
- [shop-cart](shop-cart/README.md)
- [shop-order](shop-order/README.md)
- [shop-payment](shop-payment/README.md)
- [shop-BFF](shop-BFF/README.md)
- [shop-web](shop-web/README.md)

## Локальная разработка одного сервиса

Каждый сервис — отдельный Go-модуль со своим `Makefile` и `.env.example`.  
Общий proto: `shop-proto`.  
Для полного стека удобнее `make up` из корня.
