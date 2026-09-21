-- EN: Creates per-service databases on first PostgreSQL startup (initdb.d hook).
-- RU: Создаёт отдельные базы данных для каждого сервиса при первом запуске PostgreSQL (initdb.d).

-- EN: Database for auth service (users, credentials).
-- RU: База данных сервиса auth (пользователи, учётные данные).
CREATE DATABASE auth_db;
-- EN: Database for catalog service (products, categories, stock reservations).
-- RU: База данных сервиса catalog (товары, категории, резервы остатков).
CREATE DATABASE catalog_db;
-- EN: Database for cart service (shopping cart items).
-- RU: База данных сервиса cart (позиции корзины).
CREATE DATABASE cart_db;
-- EN: Database for order service (orders, order items).
-- RU: База данных сервиса order (заказы, позиции заказов).
CREATE DATABASE order_db;
-- EN: Database for payment service (payments, transactions).
-- RU: База данных сервиса payment (платежи, транзакции).
CREATE DATABASE payment_db;
-- EN: Database for notification service (email/push outbox).
-- RU: База данных сервиса notification (outbox email/push).
CREATE DATABASE notification_db;
