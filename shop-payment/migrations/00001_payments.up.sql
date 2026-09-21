-- Create payments table and indexes for user and order lookups.
-- Создание таблицы payments и индексов для поиска по пользователю и заказу.
CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,           -- Auto-generated payment identifier / Автогенерируемый идентификатор платежа
    order_id   BIGINT NOT NULL,                 -- Reference to the originating order / Ссылка на исходный заказ
    user_id    BIGINT NOT NULL,                 -- Customer who initiated the payment / Клиент, инициировавший платёж
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount > 0),  -- Charged sum, must be positive / Списанная сумма, должна быть > 0
    status     TEXT NOT NULL,                   -- Lifecycle state: pending, success, failed / Состояние: pending, success, failed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Record insertion timestamp / Метка времени вставки записи
    CHECK (status IN ('pending', 'success', 'failed'))
);
CREATE INDEX idx_payments_user_created ON payments(user_id, created_at DESC);  -- List payments by user, newest first / Список платежей пользователя, сначала новые
CREATE INDEX idx_payments_order ON payments(order_id);  -- Lookup payment by order / Поиск платежа по заказу
