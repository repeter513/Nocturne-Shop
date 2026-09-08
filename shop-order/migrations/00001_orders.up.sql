CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    total_price NUMERIC(12, 2) NOT NULL,
    payment_id  BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'paid', 'failed', 'cancelled'))
);
CREATE TABLE order_items (
    order_id   BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity   INT NOT NULL CHECK (quantity > 0),
    name       TEXT NOT NULL,
    price      NUMERIC(12, 2) NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);
