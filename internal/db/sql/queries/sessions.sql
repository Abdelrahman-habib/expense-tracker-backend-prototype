-- name: UpsertSession :one
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
RETURNING *;

-- name: GetSession :one
SELECT * FROM "sessions"
WHERE key = $1 AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM "sessions"
WHERE key = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM "sessions"
WHERE expires_at <= CURRENT_TIMESTAMP; 