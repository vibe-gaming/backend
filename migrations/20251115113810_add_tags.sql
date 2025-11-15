-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


ALTER TABLE benefit
    ADD COLUMN tags json NOT NULL COMMENT 'Теги';


-- +goose Down
-- +goose StatementBegin
ALTER TABLE benefit
    DROP COLUMN tags;

    
SELECT 'down SQL query';
-- +goose StatementEnd
