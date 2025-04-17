-- name: GetWallet :one
SELECT * FROM wallets
WHERE wallet_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListWallets :many
SELECT * FROM wallets
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateWallet :one
INSERT INTO wallets (
    user_id,
    project_id,
    name,
    balance,
    currency,
    tags
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateWallet :one
UPDATE wallets
SET 
    name = COALESCE(sqlc.narg('name'), name),
    balance = sqlc.narg('balance'),
    currency = COALESCE(sqlc.narg('currency'), currency),
    tags = sqlc.narg('tags'),
    updated_at = CURRENT_TIMESTAMP

WHERE wallet_id = sqlc.arg('wallet_id') AND user_id = sqlc.arg('user_id')
RETURNING *;


-- name: DeleteWallet :exec
DELETE FROM wallets
WHERE wallet_id = $1 AND user_id = $2;

-- name: ListWalletsPaginated :many
SELECT * 
FROM wallets
WHERE user_id = $1 
  AND (created_at < $2 OR (created_at = $2 AND wallet_id < $3))
ORDER BY created_at DESC, wallet_id DESC
LIMIT $4;

-- name: GetProjectWallets :many
SELECT * FROM wallets
WHERE project_id = $1 AND user_id = $2
ORDER BY created_at DESC;

-- name: SearchWallets :many
SELECT *
FROM wallets
WHERE user_id = sqlc.arg('user_id')
  AND (
      sqlc.arg('name')::text = ''  -- No filter applied if sqlc.arg('name') is empty
      OR name ILIKE '%' || sqlc.arg('name') || '%'  -- Substring match
      OR name <-> sqlc.arg('name') < 0.8  -- Trigram similarity with threshold
  )
ORDER BY 
    CASE WHEN sqlc.arg('name') = '' THEN created_at END DESC,  -- If sqlc.arg('name') is empty, sort by created_at
    CASE WHEN sqlc.arg('name') <> '' THEN name <-> sqlc.arg('name') END,  -- If sqlc.arg('name') is provided, sort by trigram similarity
    length(name) ASC  -- Shorter names are preferred as tiebreaker
LIMIT sqlc.arg('limit');
