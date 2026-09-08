DELETE FROM products WHERE id > 3;
DELETE FROM categories WHERE id > 2;

UPDATE products SET description = 'Ergonomic mouse' WHERE id = 1;
UPDATE products SET description = 'RGB keyboard' WHERE id = 2;
UPDATE products SET description = 'Go programming book' WHERE id = 3;

UPDATE categories SET description = 'Gadgets and devices' WHERE id = 1;
UPDATE categories SET description = 'Physical and digital books' WHERE id = 2;
