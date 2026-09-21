-- EN: Initial catalog schema — categories, products, and stock reservations.
-- RU: Начальная схема каталога — категории, товары и резервы остатков.

-- EN: Product categories with optional parent for nesting (self-referencing FK).
-- RU: Категории товаров с опциональным родителем для вложенности (self-referencing FK).
CREATE TABLE categories (
    -- EN: Surrogate primary key, auto-incremented.
    -- RU: Суррогатный первичный ключ, автоинкремент.
    id BIGSERIAL PRIMARY KEY,
    -- EN: Display name of the category.
    -- RU: Отображаемое название категории.
    name TEXT NOT NULL,
    -- EN: Longer description of category scope; empty string by default.
    -- RU: Подробное описание области категории; пустая строка по умолчанию.
    description TEXT NOT NULL DEFAULT '',
    -- EN: Optional parent category for tree hierarchy; NULL = root category.
    -- RU: Опциональная родительская категория для дерева; NULL = корневая категория.
    parent_id BIGINT REFERENCES categories(id),
    -- EN: Timestamp when category row was created.
    -- RU: Время создания строки категории.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- EN: Catalog products with price, physical stock, and active flag for soft hide.
-- RU: Товары каталога с ценой, физическим остатком и флагом active для soft hide.
CREATE TABLE products (
    -- EN: Surrogate primary key.
    -- RU: Суррогатный первичный ключ.
    id BIGSERIAL PRIMARY KEY,
    -- EN: Product title shown in listings.
    -- RU: Название товара в каталоге.
    name TEXT NOT NULL,
    -- EN: Full product description (may contain newlines).
    -- RU: Полное описание товара (может содержать переносы строк).
    description TEXT NOT NULL DEFAULT '',
    -- EN: Unit price; must be positive (CHECK price > 0).
    -- RU: Цена за единицу; должна быть положительной (CHECK price > 0).
    price NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    -- EN: Foreign key to categories.id for filtering.
    -- RU: Внешний ключ на categories.id для фильтрации.
    category_id BIGINT REFERENCES categories(id),
    -- EN: Physical on-hand quantity; reservations do not decrement until confirm.
    -- RU: Физический остаток; резервы не уменьшают его до confirm.
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    -- EN: FALSE hides product from ListProducts queries (soft delete).
    -- RU: FALSE скрывает товар из ListProducts (soft delete).
    active BOOLEAN NOT NULL DEFAULT TRUE,
    -- EN: Row creation timestamp.
    -- RU: Время создания строки.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- EN: Last update timestamp (stock/price changes).
    -- RU: Время последнего обновления (изменения остатка/цены).
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- EN: Temporary stock holds keyed by order_id until confirm or expiry.
-- RU: Временные резервы остатков по order_id до подтверждения или истечения.
CREATE TABLE stock_reservations (
    -- EN: Reservation primary key.
    -- RU: Первичный ключ резерва.
    id BIGSERIAL PRIMARY KEY,
    -- EN: Order identifier; UNIQUE ensures one reservation row per order (idempotent reserve).
    -- RU: Идентификатор заказа; UNIQUE — одна строка резерва на заказ (идемпотентный reserve).
    order_id BIGINT NOT NULL UNIQUE,
    -- EN: JSONB map product_id (string key) → reserved quantity.
    -- RU: JSONB map product_id (строковый ключ) → зарезервированное количество.
    items JSONB NOT NULL DEFAULT '{}',
    -- EN: Lifecycle status; constrained to known values (see ReservationStatus in Go).
    -- RU: Статус жизненного цикла; ограничен известными значениями (см. ReservationStatus в Go).
    status TEXT NOT NULL DEFAULT 'active',
    -- EN: Absolute expiry time; cleanup job marks expired when expires_at <= NOW().
    -- RU: Абсолютное время истечения; cleanup помечает expired когда expires_at <= NOW().
    expires_at TIMESTAMPTZ NOT NULL,
    -- EN: When reservation row was first inserted.
    -- RU: Когда строка резерва была впервые вставлена.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('active', 'released', 'confirmed', 'expired', 'partially_released'))
);

-- EN: Partial index for category-filtered active product listings.
-- RU: Частичный индекс для выборки активных товаров по категории.
CREATE INDEX idx_products_category ON products(category_id) WHERE active = TRUE;
-- EN: Partial index for background cleanup of active reservations by expires_at.
-- RU: Частичный индекс для фоновой очистки active резервов по expires_at.
CREATE INDEX idx_reservations_active_expires ON stock_reservations(expires_at) WHERE status = 'active';
-- EN: GIN index for JSONB ? operator lookups by product_id in items map.
-- RU: GIN-индекс для JSONB ? поиска по product_id в map items.
CREATE INDEX idx_reservations_items ON stock_reservations USING GIN(items);
