-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE benefit
    ADD COLUMN views int NOT NULL DEFAULT 0 COMMENT 'Количество просмотров';

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
