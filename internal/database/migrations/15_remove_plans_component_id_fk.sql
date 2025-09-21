-- +goose Up
-- +goose StatementBegin
ALTER TABLE plans DROP FOREIGN KEY plans_ibfk_1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE plans ADD CONSTRAINT plans_ibfk_1 FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE;
-- +goose StatementEnd
