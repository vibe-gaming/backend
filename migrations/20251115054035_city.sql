-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE region (
    id BINARY(16) NOT NULL,
    name VARCHAR(255) NOT NULL COMMENT 'Название региона',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY region_idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Создание таблицы городов
CREATE TABLE city (
    id BINARY(16) NOT NULL,
    region_id BINARY(16) NOT NULL COMMENT 'ID региона',
    name VARCHAR(255) NOT NULL COMMENT 'Название города',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY city_idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE user ADD COLUMN city_id BINARY(16) DEFAULT NULL COMMENT 'ID города';

INSERT INTO region (id, name) VALUES (UUID_TO_BIN("d425ddef-9602-4e17-8f93-00e51c22bd5d"),"Республика Саха (Якутия)");

INSERT INTO city (id, name, region_id) VALUES (UUID_TO_BIN("0c0f5d68-31e3-469e-8b30-b55702659254"),"Якутск", UUID_TO_BIN("d425ddef-9602-4e17-8f93-00e51c22bd5d"));

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS city;
DROP TABLE IF EXISTS region;
ALTER TABLE user DROP COLUMN city_id;
