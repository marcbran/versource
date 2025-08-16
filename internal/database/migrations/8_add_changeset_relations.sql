-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans ADD COLUMN changeset_id INT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD CONSTRAINT fk_plans_changeset_id FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE applies ADD COLUMN changeset_id INT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE applies ADD CONSTRAINT fk_applies_changeset_id FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE applies DROP FOREIGN KEY fk_applies_changeset_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE applies DROP COLUMN changeset_id;

ALTER TABLE plans DROP FOREIGN KEY fk_plans_changeset_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP COLUMN changeset_id;
-- +goose StatementEnd
