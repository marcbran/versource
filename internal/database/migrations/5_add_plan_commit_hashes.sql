-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN merge_base VARCHAR(40) NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN head VARCHAR(40) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN merge_base;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN head;
-- +goose StatementEnd
