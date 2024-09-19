-- +goose Up
CREATE TABLE users(
id uuid primary key not null, 
created_at timestamp NOT NULL,
updated_at timestamp NOT NULL,
name text
);

-- +goose Down
drop table users;
