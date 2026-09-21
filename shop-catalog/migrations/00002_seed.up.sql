-- EN: Seed data — starter categories and products for local development and demos.
-- RU: Начальные данные — базовые категории и товары для локальной разработки и демо.

-- EN: Root categories Electronics (id=1) and Books (id=2) inserted in order.
-- RU: Корневые категории Electronics (id=1) и Books (id=2) вставляются по порядку.
INSERT INTO categories (name, description) VALUES
    ('Electronics', 'Gadgets and devices'),
    ('Books', 'Physical and digital books');

-- EN: Sample products referencing category_id 1 (Electronics) and 2 (Books).
-- RU: Примеры товаров со ссылкой на category_id 1 (Electronics) и 2 (Books).
INSERT INTO products (name, description, price, category_id, stock) VALUES
    ('Wireless Mouse', 'Ergonomic mouse', 29.99, 1, 100),
    ('Mechanical Keyboard', 'RGB keyboard', 89.99, 1, 50),
    ('Go in Action', 'Go programming book', 39.99, 2, 30);
