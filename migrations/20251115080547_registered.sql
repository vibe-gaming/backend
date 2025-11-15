-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE user ADD COLUMN registered_at DATETIME DEFAULT NULL COMMENT 'Дата регистрации';

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE user DROP COLUMN registered_at;
