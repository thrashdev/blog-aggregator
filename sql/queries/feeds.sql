-- name: CreateFeed :one
insert into feeds(id, created_at, updated_at, name, url, user_id)
values ($1, $2, $3, $4, $5, $6)
	returning *;

-- name: GetFeeds :many
select * from feeds;

-- name: GetNextFeedsToFetch :many
select * from feeds order by fetched_at NULLS FIRST LIMIT $1;

-- name: MarkFeedFetched :one
UPDATE feeds
SET fetched_at = Now(),
    updated_at = Now()
WHERE id = $1
RETURNING *;

-- name: GetFeedsAndUsernames :many
select f.id, f.name, f.url, u.name as username from feeds f 
join users u on f.user_id = u.id;

-- name: GetFeedByUrl :one
select * from feeds
where url = $1;
