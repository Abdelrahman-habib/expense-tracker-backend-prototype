-- name: CreateTag :one
INSERT INTO tags (
    user_id,
    name,
    color
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetTag :one
SELECT * FROM tags
WHERE tag_id = $1 AND user_id = $2;

-- name: ListTags :many
SELECT * FROM tags
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateTag :one
UPDATE tags
SET name = $2,
    color = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE tag_id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE tag_id = $1 AND user_id = $2;

-- name: DeleteUserTags :exec
DELETE FROM tags
WHERE user_id = $1;
