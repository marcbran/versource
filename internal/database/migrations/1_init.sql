-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source VARCHAR(255) NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS module_versions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    module_id INT NOT NULL,
    `version` VARCHAR(255) NOT NULL,
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
    variables JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (module_version_id) REFERENCES module_versions(id) ON DELETE CASCADE
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
CREATE TABLE IF NOT EXISTS changesets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    state VARCHAR(50) NOT NULL DEFAULT ('Open')
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    component_id INT NOT NULL,
    changeset_id INT NOT NULL,
    merge_base VARCHAR(255) NOT NULL,
    head VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT ('Queued'),
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX plans_component_id ON plans (component_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX plans_changeset_id ON plans (changeset_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS applies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    plan_id INT NOT NULL,
    changeset_id INT NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT ('Queued'),
    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX applies_plan_id ON applies (plan_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX applies_changeset_id ON applies (changeset_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS module_versions;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS modules;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS changesets;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS applies;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS plans;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS states;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS components;
-- +goose StatementEnd
