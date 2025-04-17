-- name: GetContact :one
SELECT * FROM contacts
WHERE contact_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListContacts :many
SELECT * FROM contacts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateContact :one
INSERT INTO contacts (
    user_id,
    name,
    phone,
    email,
    address_line1,
    address_line2,
    country,
    city,
    state_province,
    zip_postal_code,
    tags
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateContact :one
UPDATE contacts
SET 
    name = COALESCE(sqlc.narg('name'), name),
    phone = sqlc.narg('phone'),
    email = sqlc.narg('email'),
    address_line1 = sqlc.narg('address_line1'),
    address_line2 = sqlc.narg('address_line2'),
    country = sqlc.narg('country'),
    city = sqlc.narg('city'),
    state_province = sqlc.narg('state_province'),
    zip_postal_code = sqlc.narg('zip_postal_code'),
    tags = sqlc.narg('tags'),
    updated_at = CURRENT_TIMESTAMP
WHERE contact_id = sqlc.arg('contact_id') AND user_id = sqlc.arg('user_id')
RETURNING *;

-- name: DeleteContact :exec
DELETE FROM contacts
WHERE contact_id = $1 AND user_id = $2;

-- name: ListContactsPaginated :many
SELECT * 
FROM contacts
WHERE user_id = $1 
  AND (created_at < $2 OR (created_at = $2 AND contact_id < $3))
ORDER BY created_at DESC, contact_id DESC
LIMIT $4;

-- name: SearchContacts :many
SELECT *
FROM contacts
WHERE user_id = sqlc.arg('user_id')
  AND (
      sqlc.arg('name')::text = ''  -- No filter applied if sqlc.arg('name') is empty
      OR name ILIKE '%' || sqlc.arg('name') || '%'  -- Substring match
      OR name <-> sqlc.arg('name') < 0.9  -- Trigram similarity with threshold high for low sim to be included
  )
ORDER BY 
    CASE WHEN sqlc.arg('name') = '' THEN created_at END DESC,  -- If sqlc.arg('name') is empty, sort by created_at
    CASE WHEN sqlc.arg('name') <> '' THEN name <-> sqlc.arg('name') END,  -- If sqlc.arg('name') is provided, sort by trigram similarity
    length(name) ASC  -- Shorter names are preferred as tiebreaker
LIMIT sqlc.arg('limit');

-- name: SearchContactsByPhone :many
SELECT *
FROM contacts
WHERE user_id = sqlc.arg('user_id')
  AND (
      sqlc.arg('phone')::text = ''  -- No filter applied if sqlc.arg('phone') is empty
      OR phone LIKE sqlc.arg('phone') || '%'
  )
ORDER BY 
    CASE WHEN sqlc.arg('phone') = '' THEN created_at END DESC,
    CASE 
        WHEN phone = sqlc.arg('phone') THEN 1  -- Exact match
        WHEN phone LIKE sqlc.arg('phone') || '%' THEN 2  -- Starts with
        ELSE 3  -- Contains
    END,
    created_at DESC
LIMIT sqlc.arg('limit');