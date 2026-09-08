CREATE TABLE cart_items (
    user_id    BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity   INT    NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (user_id, product_id)
);
