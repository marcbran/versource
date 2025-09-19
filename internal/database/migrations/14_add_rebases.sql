-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rebases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    changeset_id INT NOT NULL,
    rebase_base VARCHAR(255) NOT NULL,
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
