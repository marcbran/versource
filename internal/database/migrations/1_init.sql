-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    `version` VARCHAR(255),
    variables JSON NOT NULL DEFAULT ('{}')
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS states (
    id INT AUTO_INCREMENT PRIMARY KEY,
    module_id INT UNIQUE NOT NULL,
    `output` JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS states;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS modules;
-- +goose StatementEnd
