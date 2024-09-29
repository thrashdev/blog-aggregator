-- name: CreatePost :one
INSERT INTO posts(id, created_at,  updated_at,  title,  url,  description,  published_at, feed_id)
VALUES(
$1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsByUser :many
select p.*, f.name as FeedName from posts p
join feeds f on p.feed_id = f.id
join feed_follows ff on ff.feed_id = f.id
join users u on ff.user_id = u.id
where u.Name = $1
ORDER BY published_at DESC
LIMIT $2;
