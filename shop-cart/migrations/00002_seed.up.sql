-- Seed sample cart items for development and testing.
-- Заполнение тестовыми позициями корзины для разработки и тестирования.
INSERT INTO cart_items (user_id, product_id, quantity) VALUES
    (1, 1, 2),  -- User 1: 2x product 1 / Пользователь 1: 2 шт. товара 1
    (1, 3, 1),  -- User 1: 1x product 3 / Пользователь 1: 1 шт. товара 3
    (2, 2, 1);  -- User 2: 1x product 2 / Пользователь 2: 1 шт. товара 2
