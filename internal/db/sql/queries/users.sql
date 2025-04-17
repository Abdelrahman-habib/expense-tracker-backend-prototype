-- name: GetUser :one
SELECT * FROM "users"
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByExternalID :one
SELECT * FROM "users"
WHERE external_id = $1 AND provider = $2 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM "users"
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO "users" (
  name,
  email,
  external_id,
  provider,
  address_line1,
  address_line2,
  country,
  city,
  state_province,
  zip_postal_code
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateUser :one
UPDATE "users"
SET 
  name = COALESCE($2, name),
  email = COALESCE($3, email),
  address_line1 = COALESCE($4, address_line1),
  address_line2 = COALESCE($5, address_line2),
  country = COALESCE($6, country),
  city = COALESCE($7, city),
  state_province = COALESCE($8, state_province),
  zip_postal_code = COALESCE($9, zip_postal_code),
  updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserRefreshToken :exec
UPDATE "users"
SET 
  refresh_token_hash = $2,
  updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE "users"
SET 
  last_login_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE FROM "users"
WHERE user_id = $1;

-- Add efficient pagination using keyset pagination
-- name: ListUsersPaginated :many
SELECT * FROM "users"
WHERE (created_at, user_id) < ($1, $2)
ORDER BY created_at DESC, user_id DESC
LIMIT $3;

-- Add efficient search
-- name: SearchUsers :many
SELECT * FROM users
WHERE name ILIKE $1
ORDER BY 
    CASE WHEN name ILIKE $1 THEN 0
         WHEN name ILIKE ($1 || '%') THEN 1
         ELSE 2
    END,
    created_at DESC
LIMIT $2;