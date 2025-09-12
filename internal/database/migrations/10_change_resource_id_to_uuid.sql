-- +goose Up
-- +goose StatementBegin
ALTER TABLE resources ADD COLUMN uuid VARCHAR(36) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources MODIFY COLUMN id INT NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources DROP PRIMARY KEY;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources ADD PRIMARY KEY (uuid);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources DROP COLUMN id;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources MODIFY COLUMN resource_id VARCHAR(36);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources ADD FOREIGN KEY (resource_id) REFERENCES resources(uuid) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET @constraint_name = (SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_NAME = 'state_resources' AND COLUMN_NAME = 'resource_id' AND REFERENCED_TABLE_NAME = 'resources' LIMIT 1);
SET @sql = CONCAT('ALTER TABLE state_resources DROP FOREIGN KEY ', @constraint_name);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources ADD COLUMN id INT AUTO_INCREMENT;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources DROP PRIMARY KEY;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources ADD PRIMARY KEY (id);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE resources DROP COLUMN uuid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources MODIFY COLUMN resource_id INT;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE state_resources ADD FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE;
-- +goose StatementEnd
