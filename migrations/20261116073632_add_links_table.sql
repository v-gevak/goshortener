-- +goose Up
CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY,
    short_key VARCHAR(11) NOT NULL UNIQUE,
    original_url TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS links;