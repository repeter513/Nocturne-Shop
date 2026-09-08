CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'success', 'failed'))
);
CREATE INDEX idx_payments_user_created ON payments(user_id, created_at DESC);
CREATE INDEX idx_payments_order ON payments(order_id);
