-- name: GetUserSettings :one
SELECT * FROM users_settings
WHERE user_id = $1 LIMIT 1;

-- name: CreateUserSettings :one
INSERT INTO users_settings (
    user_id,
    default_currency,
    default_country,
    timezone,
    date_format,
    number_format
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateUserSettings :one
UPDATE users_settings
SET 
    default_currency = COALESCE($2, default_currency),
    default_country = COALESCE($3, default_country),
    timezone = COALESCE($4, timezone),
    date_format = COALESCE($5, date_format),
    number_format = COALESCE($6, number_format),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: DeleteUserSettings :exec
DELETE FROM users_settings
WHERE user_id = $1; 