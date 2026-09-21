-- EN: Users table for auth service — stores credentials and optimistic-lock version.
-- RU: Таблица пользователей для сервиса auth — хранит учётные данные и version для optimistic locking.
CREATE TABLE users (
    -- EN: Surrogate primary key, auto-incremented.
    -- RU: Суррогатный первичный ключ, автоинкремент.
    id SERIAL PRIMARY KEY,
    -- EN: Optimistic lock counter; default 1 on insert.
    -- RU: Счётчик optimistic lock; по умолчанию 1 при вставке.
    version INT NOT NULL DEFAULT 1,
    -- EN: Unique login identifier; max 50 chars enforced by VARCHAR(50).
    -- RU: Уникальный идентификатор входа; макс. 50 символов через VARCHAR(50).
    email VARCHAR(50) NOT NULL UNIQUE,
    -- EN: bcrypt password hash; never store plaintext.
    -- RU: bcrypt-хеш пароля; plaintext не хранить.
    password_hash VARCHAR(255) NOT NULL,
    -- EN: Row creation timestamp; set once on INSERT.
    -- RU: Время создания строки; задаётся один раз при INSERT.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- EN: Last update timestamp; bump on profile/password changes.
    -- RU: Время последнего обновления; обновлять при изменении профиля/пароля.
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
