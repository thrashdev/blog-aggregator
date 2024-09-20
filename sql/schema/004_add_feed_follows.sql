-- +goose Up
create table feed_follows(
id UUID PRIMARY KEY NOT NULL,
user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
feed_id UUID REFERENCES feeds(id) NOT NULL,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL
);

-- +goose Down
drop table feed_follows;
