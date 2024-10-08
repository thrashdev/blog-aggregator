// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: feeds.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createFeed = `-- name: CreateFeed :one
insert into feeds(id, created_at, updated_at, name, url, user_id)
values ($1, $2, $3, $4, $5, $6)
	returning id, created_at, updated_at, name, url, user_id, fetched_at
`

type CreateFeedParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      sql.NullString
	Url       sql.NullString
	UserID    uuid.UUID
}

func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
	row := q.db.QueryRowContext(ctx, createFeed,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Url,
		arg.UserID,
	)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.FetchedAt,
	)
	return i, err
}

const getFeedByUrl = `-- name: GetFeedByUrl :one
select id, created_at, updated_at, name, url, user_id, fetched_at from feeds
where url = $1
`

func (q *Queries) GetFeedByUrl(ctx context.Context, url sql.NullString) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getFeedByUrl, url)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.FetchedAt,
	)
	return i, err
}

const getFeeds = `-- name: GetFeeds :many
select id, created_at, updated_at, name, url, user_id, fetched_at from feeds
`

func (q *Queries) GetFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getFeeds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
			&i.FetchedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFeedsAndUsernames = `-- name: GetFeedsAndUsernames :many
select f.id, f.name, f.url, u.name as username from feeds f 
join users u on f.user_id = u.id
`

type GetFeedsAndUsernamesRow struct {
	ID       uuid.UUID
	Name     sql.NullString
	Url      sql.NullString
	Username sql.NullString
}

func (q *Queries) GetFeedsAndUsernames(ctx context.Context) ([]GetFeedsAndUsernamesRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedsAndUsernames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedsAndUsernamesRow
	for rows.Next() {
		var i GetFeedsAndUsernamesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Url,
			&i.Username,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNextFeedsToFetch = `-- name: GetNextFeedsToFetch :many
select id, created_at, updated_at, name, url, user_id, fetched_at from feeds order by fetched_at NULLS FIRST LIMIT $1
`

func (q *Queries) GetNextFeedsToFetch(ctx context.Context, limit int32) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getNextFeedsToFetch, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
			&i.FetchedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markFeedFetched = `-- name: MarkFeedFetched :one
UPDATE feeds
SET fetched_at = Now(),
    updated_at = Now()
WHERE id = $1
RETURNING id, created_at, updated_at, name, url, user_id, fetched_at
`

func (q *Queries) MarkFeedFetched(ctx context.Context, id uuid.UUID) (Feed, error) {
	row := q.db.QueryRowContext(ctx, markFeedFetched, id)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.FetchedAt,
	)
	return i, err
}
