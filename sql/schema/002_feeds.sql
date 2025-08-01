-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;

-- Connection string : psql "postgres://postgres:postgres@localhost:5432/gator"
-- Starts database : sudo service postgresql start
-- UP Database : goose -dir ./sql/schema postgres postgres://postgres:postgres@localhost:5432/gator up