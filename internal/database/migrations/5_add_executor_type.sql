-- +goose Up
-- +goose StatementBegin
ALTER TABLE modules ADD COLUMN executor_type VARCHAR(255) NOT NULL DEFAULT 'terraform-module';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE modules DROP COLUMN executor_type;
-- +goose StatementEnd
