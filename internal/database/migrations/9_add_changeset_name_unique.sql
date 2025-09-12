-- +goose Up
-- +goose StatementBegin
ALTER TABLE changesets ADD CONSTRAINT changesets_name_unique UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE changesets DROP CONSTRAINT changesets_name_unique;
-- +goose StatementEnd
