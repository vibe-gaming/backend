-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


ALTER TABLE organization_building
    ADD COLUMN type VARCHAR(50) NOT NULL COMMENT 'Тип здания';
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE organization_building
    DROP COLUMN type;