-- Enforce one payment per order via unique index on order_id.
-- Один платёж на заказ: уникальный индекс по order_id.
-- Enables idempotent CreatePayment: duplicate inserts load the existing row.
-- Обеспечивает идемпотентность CreatePayment: повторная вставка загружает существующую строку.
CREATE UNIQUE INDEX idx_payments_order_id ON payments(order_id);
