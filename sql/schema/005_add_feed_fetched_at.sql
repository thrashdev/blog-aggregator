-- +goose Up
alter table feeds 
add column fetched_at TIMESTAMP;

-- +goose Down
alter table feeds
drop column fetched_at;
