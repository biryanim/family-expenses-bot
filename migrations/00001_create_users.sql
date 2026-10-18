-- +goose Up
CREATE TABLE users (
    telegram_id BIGINT PRIMARY KEY,
    username    TEXT        NOT NULL DEFAULT '',
    first_name  TEXT        NOT NULL DEFAULT '',
    last_name   TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

-- +goose Down
DROP TABLE users;
