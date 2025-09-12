-- +goose Up
-- +goose StatementBegin
ALTER TABLE modules ADD COLUMN name VARCHAR(255) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX modules_name_unique ON modules (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX modules_name_unique;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE modules DROP COLUMN name;
-- +goose StatementEnd
