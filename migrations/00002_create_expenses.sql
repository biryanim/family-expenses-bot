-- +goose Up
CREATE TABLE expenses (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    amount     BIGINT      NOT NULL,
    category   TEXT        NOT NULL,
    spent_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    comment    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    CONSTRAINT expenses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (telegram_id),
    CONSTRAINT expenses_amount_positive CHECK (amount > 0),
    CONSTRAINT expenses_category_valid CHECK (
        category IN ('food', 'transport', 'housing', 'entertainment', 'health', 'shopping', 'other')
    )
);

CREATE INDEX expenses_user_id_spent_at_idx ON expenses (user_id, spent_at DESC);
CREATE INDEX expenses_spent_at_idx ON expenses (spent_at DESC);

-- +goose Down
DROP TABLE expenses;
