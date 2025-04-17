-- +goose Up
-- +goose StatementBegin
CREATE INDEX project_name_trgm ON projects USING gin (name gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS project_name_trgm;
-- +goose StatementEnd
