-- +goose Up
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS merges;
-- +goose StatementEnd
