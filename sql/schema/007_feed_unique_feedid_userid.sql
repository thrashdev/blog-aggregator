-- +goose Up
-- alter table feed_follows
-- alter feed_id TYPE UUID REFERENCES feeds(id) ON DELETE CASCADE NOT NULL;

alter table feed_follows
add constraint unique_feed_user_combination UNIQUE(feed_id, user_id);

-- +goose Down
-- alter table feed_follows
-- alter feed_id TYPE UUID REFERENCES feeds(id) NOT NULL;

alter table feed_follows
drop constraint unique_feed_user_combination;
