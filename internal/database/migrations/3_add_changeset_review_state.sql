-- +goose Up
-- +goose StatementBegin
ALTER TABLE changesets ADD COLUMN review_state VARCHAR(50) NOT NULL DEFAULT ('Draft');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE changesets DROP COLUMN review_state;
-- +goose StatementEnd
