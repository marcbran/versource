-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN state VARCHAR(20) NOT NULL DEFAULT 'Queued' CHECK (state IN ('Queued', 'Started', 'Aborted', 'Completed', 'Failed', 'Cancelled'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE applies ADD COLUMN state VARCHAR(20) NOT NULL DEFAULT 'Queued' CHECK (state IN ('Queued', 'Started', 'Aborted', 'Completed', 'Failed', 'Cancelled'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN state;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE applies DROP COLUMN state;
-- +goose StatementEnd
