-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE benefit
    ADD COLUMN category VARCHAR(50) DEFAULT NULL COMMENT 'Категория льготы (medicine, transport, food, clothing, other)';

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE benefit
    DROP COLUMN category;
