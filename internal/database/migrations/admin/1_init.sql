-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS changesets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    state VARCHAR(50) NOT NULL DEFAULT ('Open'),
    review_state VARCHAR(50) NOT NULL DEFAULT ('Draft')
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    component_id INT NOT NULL,
    changeset_id INT NOT NULL,
    `from` VARCHAR(255) NULL,
    `to` VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT ('Queued'),
    `add` INT NULL,
    `change` INT NULL,
    `destroy` INT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX plans_component_id ON plans (component_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX plans_changeset_id ON plans (changeset_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX plans_state ON plans (state);
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

-- +goose StatementBegin
CREATE INDEX applies_state ON applies (state);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS merges (
    id INT AUTO_INCREMENT PRIMARY KEY,
    changeset_id INT NOT NULL,
    merge_base VARCHAR(255) NOT NULL,
    head VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT ('Queued'),
    FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX merges_changeset_id ON merges (changeset_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX merges_state ON merges (state);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rebases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    changeset_id INT NOT NULL,
    merge_base VARCHAR(255) NOT NULL,
    head VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT ('Queued'),
    FOREIGN KEY (changeset_id) REFERENCES changesets(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX rebases_changeset_id ON rebases (changeset_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX rebases_state ON rebases (state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rebases;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS merges;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS applies;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS plans;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS changesets;
-- +goose StatementEnd
