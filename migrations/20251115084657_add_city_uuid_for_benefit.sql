-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

alter table benefit
    add city_id BINARY(16) DEFAULT NULL;


UPDATE benefit SET city_id = UUID_TO_BIN("0c0f5d68-31e3-469e-8b30-b55702659254");

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
