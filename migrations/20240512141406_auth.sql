-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE user (
    id BINARY(16) NOT NULL,
    login VARCHAR(255) UNIQUE COMMENT 'Ник пользователя',
    password VARCHAR(255) COMMENT 'Пароль',
    email VARCHAR(255) UNIQUE COMMENT 'Email пользователя',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE refresh_session (
    id BINARY(16) NOT NULL,
    user_id BINARY(16) NOT NULL,
    refresh_token BINARY(16) NOT NULL,
    user_agent VARCHAR(255) NOT NULL COMMENT 'User Agent пользователя',
    ip varchar(15) NOT NULL COMMENT 'IP пользователя',
    expires_in DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE user;
DROP TABLE refresh_session;