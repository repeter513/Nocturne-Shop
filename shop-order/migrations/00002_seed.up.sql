INSERT INTO orders (id, user_id, status, total_price) VALUES
    (1, 1, 'paid', 99.97),
    (2, 2, 'paid', 89.99);

SELECT setval('orders_id_seq', (SELECT MAX(id) FROM orders));

INSERT INTO order_items (order_id, product_id, quantity, name, price) VALUES
    (1, 1, 2, 'Wireless Mouse', 29.99),
    (1, 3, 1, 'Go in Action', 39.99),
    (2, 2, 1, 'Mechanical Keyboard', 89.99);
