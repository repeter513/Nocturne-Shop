-- Migration: seed sample orders for development.
-- Миграция: начальные данные заказов для разработки.

-- Insert sample paid orders.
-- Вставка примеров оплаченных заказов.
INSERT INTO orders (id, user_id, status, total_price) VALUES
    (1, 1, 'paid', 99.97),   -- User 1 order / Заказ пользователя 1
    (2, 2, 'paid', 89.99);   -- User 2 order / Заказ пользователя 2

-- Sync orders_id_seq with the highest inserted ID.
-- Синхронизация orders_id_seq с максимальным вставленным ID.
SELECT setval('orders_id_seq', (SELECT MAX(id) FROM orders));

-- Insert line items for seeded orders.
-- Вставка позиций для начальных заказов.
INSERT INTO order_items (order_id, product_id, quantity, name, price) VALUES
    (1, 1, 2, 'Wireless Mouse', 29.99),       -- Order 1, product 1 / Заказ 1, товар 1
    (1, 3, 1, 'Go in Action', 39.99),         -- Order 1, product 3 / Заказ 1, товар 3
    (2, 2, 1, 'Mechanical Keyboard', 89.99); -- Order 2, product 2 / Заказ 2, товар 2
