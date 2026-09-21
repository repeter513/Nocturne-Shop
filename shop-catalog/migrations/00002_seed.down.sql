-- EN: Remove all seed categories and products (destructive rollback of 00002_seed).
-- RU: Удаление всех seed категорий и товаров (деструктивный откат 00002_seed).

DELETE FROM products;
DELETE FROM categories;
