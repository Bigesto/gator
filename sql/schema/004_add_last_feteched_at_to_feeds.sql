-- +goose Up
ALTER TABLE feeds
ADD last_fetched_at TIMESTAMP;

-- +goose Down
ALTER TABLE feeds
DROP COLUMN last_fetched_at;

-- Connection string : psql "postgres://postgres:postgres@localhost:5432/gator"
-- Starts database : sudo service postgresql start
-- UP Database : goose -dir ./sql/schema postgres postgres://postgres:postgres@localhost:5432/gator up