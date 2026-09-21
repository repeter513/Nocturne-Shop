-- EN: Rollback initial catalog schema — drops reservations, products, categories in FK order.
-- RU: Откат начальной схемы каталога — удаление reservations, products, categories в порядке FK.

DROP TABLE IF EXISTS stock_reservations;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
