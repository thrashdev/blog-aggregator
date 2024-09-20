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
