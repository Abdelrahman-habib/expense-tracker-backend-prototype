-- +goose Up
CREATE TABLE "sessions" (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL,
    value BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create an index on the key for faster lookups
CREATE UNIQUE INDEX sessions_key_idx ON sessions(key);

-- +goose Down
DROP INDEX IF EXISTS sessions_key_idx;
DROP TABLE IF EXISTS "sessions"; 