-- EN: Rollback catalog enrichment — remove products/categories added in 00003, restore seed descriptions.
-- RU: Откат расширения каталога — удаление product/category из 00003, восстановление seed-описаний.

-- EN: Delete products inserted by enrichment migration (id > 3).
-- RU: Удаление товаров, вставленных миграцией enrichment (id > 3).
DELETE FROM products WHERE id > 3;
-- EN: Delete categories added in enrichment (id > 2).
-- RU: Удаление категорий, добавленных в enrichment (id > 2).
DELETE FROM categories WHERE id > 2;

-- EN: Restore original short English product descriptions from seed migration.
-- RU: Восстановление коротких английских описаний товаров из seed-миграции.
UPDATE products SET description = 'Ergonomic mouse' WHERE id = 1;
UPDATE products SET description = 'RGB keyboard' WHERE id = 2;
UPDATE products SET description = 'Go programming book' WHERE id = 3;

-- EN: Restore original English category descriptions from seed migration.
-- RU: Восстановление английских описаний категорий из seed-миграции.
UPDATE categories SET description = 'Gadgets and devices' WHERE id = 1;
UPDATE categories SET description = 'Physical and digital books' WHERE id = 2;
