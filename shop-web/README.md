# shop-web

Фронт магазина **Nocturne**: каталог, корзина, заказы, оплата.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](README.md)

В compose **нет порта на хост** — статика отдаётся через [Envoy](../shop-infra/envoy/envoy.yaml) на http://localhost:8080.

## Стек

- React 19 + TypeScript
- Vite 6
- react-router-dom 7
- nginx (только статика в Docker)

## Страницы

| Путь | Описание |
|------|----------|
| `/` | каталог |
| `/products/:id` | товар |
| `/cart` | корзина → `CreateOrder` |
| `/orders` | список заказов |
| `/orders/:id` | заказ: оплата / отмена |
| `/login` | вход |
| `/register` | регистрация |

## Быстрый старт

### Через compose

```bash
cd ../shop-infra && make up
```

Открыть http://localhost:8080 — Envoy маршрутизирует `/` сюда, `/api/` → BFF.

### Локально (dev)

Нужен поднятый стек (`make up`) — Vite проксирует `/api/v1` на Envoy `:8080`.

```bash
cp .env.example .env
npm install
npm run dev
```

Dev-сервер: http://localhost:3000 (API через прокси на `:8080`).

| Переменная | Описание |
|------------|----------|
| `VITE_API_BASE_URL` | базовый путь API (default `/api/v1`) |

## Скрипты

| Команда | Описание |
|---------|----------|
| `npm run dev` | Vite на `:3000` |
| `npm run build` | сборка в `dist/` |
| `npm run preview` | превью сборки |
