-- Migration: create orders and order_items tables.
-- Миграция: создание таблиц orders и order_items.

-- Orders table stores purchase aggregates.
-- Таблица orders хранит агрегаты покупок.
CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,              -- Auto-generated order identifier / Автогенерируемый идентификатор заказа
    user_id     BIGINT NOT NULL,                    -- Customer who placed the order / Клиент, оформивший заказ
    status      TEXT NOT NULL DEFAULT 'pending',    -- Lifecycle state: pending, paid, failed, cancelled / Состояние: pending, paid, failed, cancelled
    total_price NUMERIC(12, 2) NOT NULL,            -- Sum of all line items at order time / Сумма всех позиций на момент заказа
    payment_id  BIGINT,                             -- Reference to payment record once paid / Ссылка на платёж после оплаты
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Order creation timestamp / Метка времени создания заказа
    CHECK (status IN ('pending', 'paid', 'failed', 'cancelled'))
);

-- Order items table stores line items for each order.
-- Таблица order_items хранит позиции каждого заказа.
CREATE TABLE order_items (
    order_id   BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,  -- Parent order / Родительский заказ
    product_id BIGINT NOT NULL,                                          -- Catalog product reference / Ссылка на товар в каталоге
    quantity   INT NOT NULL CHECK (quantity > 0),                        -- Units ordered, must be positive / Заказанное количество, > 0
    name       TEXT NOT NULL,                                            -- Product name snapshot / Снимок названия товара
    price      NUMERIC(12, 2) NOT NULL,                                  -- Unit price snapshot / Снимок цены за единицу
    PRIMARY KEY (order_id, product_id)                                   -- One row per product per order / Одна строка на товар в заказе
);

-- Index for listing user orders by creation time.
-- Индекс для выборки заказов пользователя по времени создания.
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);
