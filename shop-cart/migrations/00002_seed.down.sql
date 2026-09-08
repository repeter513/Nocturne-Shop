DELETE FROM cart_items
WHERE (user_id, product_id) IN (
    (1, 1),
    (1, 3),
    (2, 2)
);
