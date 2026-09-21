-- Drop payments indexes and table.
-- Удаление индексов и таблицы payments.
DROP INDEX IF EXISTS idx_payments_order;
DROP INDEX IF EXISTS idx_payments_user_created;
DROP TABLE IF EXISTS payments;
