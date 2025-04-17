-- +goose Up
ALTER TABLE "users" 
  RENAME COLUMN clerk_ex_user_id TO external_id;

ALTER TABLE "users"
  ADD COLUMN provider VARCHAR(20) NOT NULL DEFAULT 'clerk',
  ADD COLUMN refresh_token_hash TEXT,
  ADD COLUMN last_login_at TIMESTAMP;

-- Create a unique constraint on external_id and provider
CREATE UNIQUE INDEX users_external_id_provider_idx ON users(external_id, provider);

-- +goose Down
DROP INDEX IF EXISTS users_external_id_provider_idx;

ALTER TABLE "users"
  DROP COLUMN provider,
  DROP COLUMN refresh_token_hash,
  DROP COLUMN last_login_at;

ALTER TABLE "users"
  RENAME COLUMN external_id TO clerk_ex_user_id; 