-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


CREATE TABLE favorite (
    id BINARY(16) NOT NULL,
    user_id BINARY(16) NOT NULL,
    benefit_id BINARY(16) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id)
);


-- +goose Down
-- +goose StatementBegin
DROP TABLE favorite;
-- +goose StatementEnd
