-- +goose Up
-- +goose StatementBegin
ALTER TABLE module_versions ADD COLUMN variables JSON DEFAULT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE module_versions ADD COLUMN outputs JSON DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE module_versions DROP COLUMN outputs;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE module_versions DROP COLUMN variables;
-- +goose StatementEnd
