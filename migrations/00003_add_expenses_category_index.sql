-- +goose Up
CREATE INDEX expenses_category_spent_at_idx ON expenses (category, spent_at DESC);

-- +goose Down
DROP INDEX expenses_category_spent_at_idx;
