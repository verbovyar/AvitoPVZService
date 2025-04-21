-- +goose Up
-- +goose StatementBegin
-- создаём отдельную схему для проекта
CREATE SCHEMA IF NOT EXISTS avito_schema;

-- таблица пользователей
CREATE TABLE avito_schema.users (
                                    id UUID PRIMARY KEY,
                                    email TEXT NOT NULL UNIQUE,
                                    password_hash TEXT NOT NULL,
                                    role TEXT NOT NULL,
                                    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- таблица ПВЗ
CREATE TABLE avito_schema.pvz (
                                  id UUID PRIMARY KEY,
                                  city TEXT NOT NULL,
                                  registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                  is_reception_open BOOLEAN NOT NULL DEFAULT FALSE,
                                  receptions JSONB NOT NULL DEFAULT '[]'
);

-- индекс для ускорения поиска по email
CREATE INDEX IF NOT EXISTS idx_users_email ON avito_schema.users(email);

-- GIN‑индекс для быстрого поиска по JSONB receptions
CREATE INDEX IF NOT EXISTS idx_pvz_receptions ON avito_schema.pvz USING GIN (receptions);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS avito_schema.pvz;
DROP TABLE IF EXISTS avito_schema.users;
DROP SCHEMA IF EXISTS avito_schema;
-- +goose StatementEnd
