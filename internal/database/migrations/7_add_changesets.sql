-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS changesets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL DEFAULT 'Draft',
    INDEX idx_changesets_name (name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS changesets;
-- +goose StatementEnd
