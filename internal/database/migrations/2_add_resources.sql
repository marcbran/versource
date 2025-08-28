-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    state_id INT NOT NULL,
    address VARCHAR(500) NOT NULL,
    mode VARCHAR(50) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    count INT NULL,
    for_each VARCHAR(255) NULL,
    type VARCHAR(255) NOT NULL,
    attributes JSON NOT NULL DEFAULT ('{}'),
    FOREIGN KEY (state_id) REFERENCES states(id) ON DELETE CASCADE,
    UNIQUE KEY unique_state_address (state_id, address)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX resources_state_id ON resources (state_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX resources_type ON resources (type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS resources;
-- +goose StatementEnd
