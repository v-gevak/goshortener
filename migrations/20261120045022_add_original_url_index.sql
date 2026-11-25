-- +goose Up
CREATE UNIQUE INDEX original_url_idx ON links(original_url);

-- +goose Down
DROP INDEX original_url_idx;
