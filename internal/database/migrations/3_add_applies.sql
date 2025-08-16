-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS applies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    plan_id INT UNIQUE NOT NULL,
    FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS applies;
-- +goose StatementEnd
