-- +goose Up
-- +goose StatementBegin
ALTER TABLE components ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT ('Ready');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE components DROP COLUMN status;
-- +goose StatementEnd
