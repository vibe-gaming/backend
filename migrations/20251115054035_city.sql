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

-- Вставка региона Республика Саха (Якутия)
INSERT INTO region (id, name) VALUES (uuid_to_bin('019a8614-33dd-7dba-b813-bd1d37c0ff0f'), 'Республика Саха (Якутия)');

-- Вставка города Якутск
INSERT INTO city (id, region_id, name) VALUES (uuid_to_bin('019a8614-6970-77b1-ba23-768f104e3124'), uuid_to_bin('019a8614-33dd-7dba-b813-bd1d37c0ff0f'), 'Якутск');

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS city;
DROP TABLE IF EXISTS region;
ALTER TABLE user DROP COLUMN city_id;
