-- Migration rollback: remove seeded orders.
-- Откат миграции: удаление начальных заказов.

DELETE FROM orders
WHERE id IN (1, 2);
