-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE user ADD COLUMN group_type JSON DEFAULT NULL COMMENT 'Тип группы пользователя';

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE user DROP COLUMN group_type;