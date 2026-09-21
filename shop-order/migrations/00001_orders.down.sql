-- Migration rollback: drop orders schema objects.
-- Откат миграции: удаление объектов схемы заказов.

DROP INDEX IF EXISTS idx_orders_user_created;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
