-- +goose Up
-- +goose StatementBegin
RENAME TABLE resources TO state_resources;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources ADD COLUMN resource_id INT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    provider VARCHAR(255) NOT NULL,
    provider_alias VARCHAR(255) NULL,
    resource_type VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NULL,
    name VARCHAR(255) NOT NULL,
    attributes JSON NOT NULL DEFAULT ('{}'),
    UNIQUE KEY unique_resource (provider, provider_alias, resource_type, namespace, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX resources_provider ON resources (provider);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX resources_provider_type ON resources (provider, resource_type);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX resources_provider_alias ON resources (provider, provider_alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS resources;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources DROP COLUMN resource_id;
-- +goose StatementEnd

-- +goose StatementBegin
RENAME TABLE state_resources TO resources;
-- +goose StatementEnd
