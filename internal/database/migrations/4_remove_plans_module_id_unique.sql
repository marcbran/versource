-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans DROP FOREIGN KEY plans_ibfk_1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP INDEX module_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD INDEX module_id (module_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD CONSTRAINT plans_ibfk_1 FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans DROP FOREIGN KEY plans_ibfk_1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans DROP INDEX module_id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD UNIQUE INDEX module_id (module_id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE plans ADD CONSTRAINT plans_ibfk_1 FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE;
-- +goose StatementEnd
