-- +goose Up
ALTER TABLE users
ADD CONSTRAINT users_name_unique UNIQUE(name);

-- +goose Down
ALTER TABLE users
DROP CONSTRAINT users_name_unique;
