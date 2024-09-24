-- name: CreateUser :one
insert into users(id, created_at, updated_at, name, apikey)
values ($1, $2, $3, $4, encode(sha256(random()::text::bytea), 'hex'))
RETURNING *;

-- name: GetUserByApiKey :one
select * from users
where apikey = $1;

-- name: GetUserByName :one
select * from users
where name = $1;

-- name: GetUsers :many
select * from users;

-- name: DeleteAllUsers :exec
delete from users;
