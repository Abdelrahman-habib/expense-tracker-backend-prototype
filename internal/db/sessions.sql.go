// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sessions.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteExpiredSessions = `-- name: DeleteExpiredSessions :exec
DELETE FROM "sessions"
WHERE expires_at <= CURRENT_TIMESTAMP
`

func (q *Queries) DeleteExpiredSessions(ctx context.Context) error {
	_, err := q.db.Exec(ctx, deleteExpiredSessions)
	return err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE FROM "sessions"
WHERE key = $1
`

func (q *Queries) DeleteSession(ctx context.Context, key string) error {
	_, err := q.db.Exec(ctx, deleteSession, key)
	return err
}

const getSession = `-- name: GetSession :one
SELECT session_id, key, value, expires_at, created_at, updated_at FROM "sessions"
WHERE key = $1 AND expires_at > CURRENT_TIMESTAMP
LIMIT 1
`

func (q *Queries) GetSession(ctx context.Context, key string) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, key)
	var i Session
	err := row.Scan(
		&i.SessionID,
		&i.Key,
		&i.Value,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const upsertSession = `-- name: UpsertSession :one
INSERT INTO "sessions" (
    key,
    value,
    expires_at
) VALUES (
    $1, $2, $3
)
ON CONFLICT (key) DO UPDATE
SET 
    value = EXCLUDED.value,
    expires_at = EXCLUDED.expires_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING session_id, key, value, expires_at, created_at, updated_at
`

type UpsertSessionParams struct {
	Key       string           `json:"key"`
	Value     []byte           `json:"value"`
	ExpiresAt pgtype.Timestamp `json:"expiresAt"`
}

func (q *Queries) UpsertSession(ctx context.Context, arg UpsertSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, upsertSession, arg.Key, arg.Value, arg.ExpiresAt)
	var i Session
	err := row.Scan(
		&i.SessionID,
		&i.Key,
		&i.Value,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
