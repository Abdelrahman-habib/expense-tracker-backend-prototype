-- +goose Up
-- +goose StatementBegin
CREATE INDEX contact_name_trgm ON contacts USING gin (name gin_trgm_ops);
CREATE INDEX idx_contacts_phone ON contacts (phone);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS contact_name_trgm;
DROP INDEX IF EXISTS idx_contacts_phone;
-- +goose StatementEnd
