-- name: GetProject :one
SELECT * FROM projects
WHERE project_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (
    user_id,
    name,
    description,
    status,
    start_date,
    end_date,
    budget,
    actual_cost,
    address_line1,
    address_line2,
    country,
    city,
    state_province,
    zip_postal_code,
    website,
    tags
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = sqlc.narg('description'),
    status = COALESCE(sqlc.narg('status'), status),
    start_date = sqlc.narg('start_date'),
    end_date = sqlc.narg('end_date'),
    budget = sqlc.narg('budget'),
    address_line1 = sqlc.narg('address_line1'),
    address_line2 = sqlc.narg('address_line2'),
    country = sqlc.narg('country'),
    city = sqlc.narg('city'),
    state_province = sqlc.narg('state_province'),
    zip_postal_code = sqlc.narg('zip_postal_code'),
    website = sqlc.narg('website'),
    tags = sqlc.narg('tags'),
    updated_at = CURRENT_TIMESTAMP
WHERE 
    project_id = sqlc.arg('project_id')
    AND user_id = sqlc.arg('user_id')
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE project_id = $1 AND user_id = $2;

-- name: ListProjectsPaginated :many
SELECT *
FROM projects
WHERE user_id = $1 
  AND (created_at < $2 OR (created_at = $2 AND project_id < $3))
ORDER BY created_at DESC, project_id DESC
LIMIT $4;

-- name: SearchProjects :many
SELECT * FROM projects
WHERE user_id = sqlc.arg('user_id') 
  AND (sqlc.arg('name')::text = '' OR (
    name <-> sqlc.arg('name') < 0.8 OR  
    name ILIKE '%' || sqlc.arg('name') || '%'  
  ))
ORDER BY 
    CASE WHEN sqlc.arg('name') = '' THEN created_at END DESC,  -- If sqlc.arg('name') is empty, sort by created_at
    CASE WHEN sqlc.arg('name') <> '' THEN name <-> sqlc.arg('name') END,  -- If sqlc.arg('name') is provided, sort by trigram similarity
    length(name) ASC  -- Shorter names are preferred as tiebreaker
LIMIT sqlc.arg('limit');