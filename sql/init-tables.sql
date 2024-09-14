DROP TABLE users;
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY, 
    email VARCHAR(255) NOT NULL, 
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(10),
    code_expiration TIMESTAMPTZ,
    UNIQUE (email, phone)  -- Добавлено уникальное ограничение
);
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
