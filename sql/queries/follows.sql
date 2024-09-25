-- name: CreateFeedFollow :one
INSERT INTO feed_follows(id, user_id, feed_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
	RETURNING *;

-- name: GetFeedFollowsByUserID :many
select * from feed_follows
where user_id = $1;

-- name: DeleteFeedFollowsByID :exec
DELETE FROM feed_follows 
WHERE id = $1;

-- name: CreateFeedFollowCLI :one
WITH inserted_follow as (
	INSERT INTO feed_follows(id, user_id, feed_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
		returning *
)

select if.*, u.name, f.name from inserted_follow as if
join users u on if.user_id = u.id
join feeds f on if.feed_id = f.id;

-- name: GetFeedFollowsByUserIDCLI :many
select ff.*, u.name, f.name from feed_follows ff
join users u on ff.user_id = u.id
join feeds f on ff.feed_id = f.id
where u.id = $1;
