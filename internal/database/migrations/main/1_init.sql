-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    source VARCHAR(255) NOT NULL UNIQUE,
    executor_type VARCHAR(255) NOT NULL DEFAULT 'terraform-jsonnet'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS module_versions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    module_id INT NOT NULL,
    `version` VARCHAR(255) NOT NULL,
    variables JSON DEFAULT NULL,
    outputs JSON DEFAULT NULL,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX module_versions_module_id ON module_versions (module_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS components (
    id INT AUTO_INCREMENT PRIMARY KEY,
    module_version_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT ('Ready'),
    variables JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (module_version_id) REFERENCES module_versions(id) ON DELETE CASCADE,
    UNIQUE KEY unique_component (module_version_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX components_module_version_id ON components (module_version_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS states (
    id INT AUTO_INCREMENT PRIMARY KEY,
    component_id INT UNIQUE NOT NULL,
    `output` JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX states_component_id ON states (component_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS resources (
    uuid VARCHAR(36) PRIMARY KEY,
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

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS state_resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    state_id INT NOT NULL,
    resource_id VARCHAR(36) NULL,
    address VARCHAR(500) NOT NULL,
    mode VARCHAR(50) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    count INT NULL,
    for_each VARCHAR(255) NULL,
    type VARCHAR(255) NOT NULL,
    attributes JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (state_id) REFERENCES states(id) ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES resources(uuid) ON DELETE CASCADE,
    UNIQUE KEY unique_state_resource (state_id, address)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX state_resources_state_id ON state_resources (state_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX state_resources_type ON state_resources (type);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS view_resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    query TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS view_resources;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS state_resources;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS resources;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS states;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS components;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS module_versions;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS modules;
-- +goose StatementEnd
