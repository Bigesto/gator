-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;

-- Connection string : psql "postgres://postgres:postgres@localhost:5432/gator"
-- Starts database : sudo service postgresql start
-- UP Database : goose -dir ./sql/schema postgres postgres://postgres:postgres@localhost:5432/gator up