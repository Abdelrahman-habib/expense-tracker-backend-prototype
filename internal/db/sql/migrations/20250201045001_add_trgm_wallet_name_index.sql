-- +goose Up
-- +goose StatementBegin
CREATE INDEX wallet_name_trgm ON wallets USING gin (name gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS wallet_name_trgm;
-- +goose StatementEnd
