-- +goose Up
-- +goose StatementBegin
ALTER TABLE components ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT ('');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX components_name_unique ON components(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS components_name_unique;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE components DROP COLUMN name;
-- +goose StatementEnd
