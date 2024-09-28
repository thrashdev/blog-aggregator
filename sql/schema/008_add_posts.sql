-- +goose Up
CREATE TABLE posts(
id UUID NOT NULL,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP,
title TEXT NOT NULL,
url TEXT NOT NULL UNIQUE,
description TEXT,
published_at TIMESTAMP,
feed_id UUID REFERENCES feeds(id) NOT NULL
);



-- +goose Down
DROP TABLE posts;
