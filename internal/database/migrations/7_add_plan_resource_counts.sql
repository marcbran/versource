-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN `add` INT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN `change` INT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN `destroy` INT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN `add`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN `change`;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN `destroy`;
-- +goose StatementEnd
