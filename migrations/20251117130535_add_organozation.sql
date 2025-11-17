-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


CREATE TABLE organization (
    id BINARY(16) NOT NULL,
    name VARCHAR(255) NOT NULL COMMENT 'Название организации',
    description TEXT NOT NULL COMMENT 'Описание организации',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE organization_building (
    id BINARY(16) NOT NULL,
    organization_id BINARY(16) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    address VARCHAR(255) NOT NULL COMMENT 'Адрес организации',
    latitude FLOAT NOT NULL COMMENT 'Широта',
    longitude FLOAT NOT NULL COMMENT 'Долгота',
    phone_number VARCHAR(20) NOT NULL COMMENT 'Телефон организации',
    start_time DATETIME NOT NULL COMMENT 'Время начала работы',
    end_time DATETIME NOT NULL COMMENT 'Время окончания работы',
    is_open BOOLEAN NOT NULL COMMENT 'Открыта ли организация',
    tags JSON NOT NULL COMMENT 'Теги организации',
    PRIMARY KEY (id)
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd


DROP TABLE organization;
DROP TABLE organization_building;