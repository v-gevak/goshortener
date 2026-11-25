-- +goose Up
ALTER TABLE links ADD COLUMN user_id VARCHAR(32);

-- +goose Down
ALTER TABLE links DROP COLUMN user_id;
