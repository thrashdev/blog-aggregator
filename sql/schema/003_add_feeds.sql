-- +goose Up
create table feeds(
	id uuid primary key, 
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	name text,
	url text , 
	user_id uuid references users(id) on delete cascade,
	constraint url_unique unique (url)
);

-- +goose Down
drop table feeds;
