# shop-web

Фронт магазина **Nocturne**: каталог, корзина, заказы, оплата.

**Экосистема:** [infra](../shop-infra/README.md) · [proto](../shop-proto/README.md) · [auth](../shop-auth/README.md) · [catalog](../shop-catalog/README.md) · [cart](../shop-cart/README.md) · [order](../shop-order/README.md) · [payment](../shop-payment/README.md) · [bff](../shop-BFF/README.md) · [web](README.md)

API: HTTP [shop-BFF](../shop-BFF/README.md) (`/api/v1`).  
Локальный стек: [shop-infra](../shop-infra/README.md) (порт `3000`).

## Стек

- React 19 + TypeScript
- Vite 6
- react-router-dom 7

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

Открыть http://localhost:3000. Nginx проксирует `/api/` на BFF внутри docker-сети (`bff:8080`).  
Проверка BFF с хоста: http://localhost:8090/health

### Локально (dev)

Нужен запущенный BFF. Vite проксирует `/api/v1` на `http://localhost:8090` — тот же порт, что `BFF_HTTP_PORT` в compose.

```bash
cp .env.example .env
npm install
npm run dev
```

| Переменная | Описание |
|------------|----------|
| `VITE_API_BASE_URL` | базовый путь API (default `/api/v1`) |

## Скрипты

| Команда | Описание |
|---------|----------|
| `npm run dev` | Vite на `:3000` |
| `npm run build` | сборка в `dist/` |
| `npm run preview` | превью сборки |
