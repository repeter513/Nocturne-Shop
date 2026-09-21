-- Create cart_items table storing per-user product quantities.
-- Создание таблицы cart_items с количеством товаров для каждого пользователя.
CREATE TABLE cart_items (
    user_id    BIGINT NOT NULL,  -- Owner user identifier / Идентификатор пользователя-владельца
    product_id BIGINT NOT NULL,  -- Catalog product reference / Ссылка на товар в каталоге
    quantity   INT    NOT NULL CHECK (quantity > 0),  -- Units in cart, must be positive / Количество в корзине, должно быть > 0
    PRIMARY KEY (user_id, product_id)  -- One row per user+product pair / Одна строка на пару пользователь+товар
);
