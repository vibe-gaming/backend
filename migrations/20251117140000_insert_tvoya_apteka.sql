-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


ALTER TABLE benefit
    ADD COLUMN organization_id BINARY(16) DEFAULT NULL COMMENT 'ID организации';

ALTER TABLE benefit
    ADD COLUMN category VARCHAR(50) DEFAULT NULL COMMENT 'Категория льготы';

-- Вставка организации "Твоя Аптека" для Благовещенска
INSERT INTO organization (id, name, description, created_at, updated_at, deleted_at)
VALUES (
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'Твоя Аптека',
    'Аптечная сеть в Благовещенске. Круглосуточная аптечная служба заказов: +7(914) 555-55-55. Широкий ассортимент лекарств, медицинских изделий, косметики и товаров для здоровья.',
    NOW(),
    NOW(),
    NULL
);

-- Вставка аптек организации в Благовещенске
INSERT INTO organization_building (id, organization_id, address, latitude, longitude, phone_number, start_time, end_time, is_open, tags, created_at, updated_at, deleted_at)
VALUES
-- Аптека 1: Ул. 50 лет Октября, 27а
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. 50 лет Октября, 27а',
    50.2670, 127.5270,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 2: Ул. Кантемирова, 1
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Кантемирова, 1',
    50.2705, 127.5360,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 3: Ул. Театральная, 23
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Театральная, 23',
    50.2820, 127.5170,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 4: Игнатьевское шоссе, 14/6
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, Игнатьевское шоссе, 14/6',
    50.2490, 127.5510,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 22:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 5: Ул. Театральная, 170 (в ТЦ Реал)
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Театральная, 170 (в ТЦ Реал)',
    50.2350, 127.5450,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 6: Ул. Ленина, 113
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Ленина, 113',
    50.2890, 127.5320,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 7: с. Чигири, ул. Василенко, 1
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, с. Чигири, ул. Василенко, 1',
    50.2150, 127.4980,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 21:45:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 8: Ул. Красноармейская, 153, ТЦ Фестиваль парк
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Красноармейская, 153, ТЦ Фестиваль парк',
    50.2780, 127.5420,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 9: с. Чигири, ул. Садовая, 16
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, с. Чигири, ул. Садовая, 16',
    50.2120, 127.5020,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 22:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 10: Игнатьевское шоссе, 25/14
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, Игнатьевское шоссе, 25/14',
    50.2460, 127.5580,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 11: Ул. Амурская, 311, ТЦ Авоська
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    'г. Благовещенск, ул. Амурская, 311, ТЦ Авоська',
    50.2940, 127.5180,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
);

-- Вставка benefit для аптечной сети "Твоя Аптека"
INSERT INTO benefit (
    id, title, description, valid_from, valid_to, type,
    target_group_ids, longitude, latitude, region, city_id,
    category, requirment, how_to_use, source_url, tags, views,
    organization_id, created_at, updated_at, deleted_at
)
VALUES (
    UNHEX(REPLACE(UUID(), '-', '')),
    'Скидка 10% для пенсионеров и инвалидов в сети аптек "Твоя Аптека"',
    'Аптечная сеть "Твоя Аптека" предоставляет постоянную скидку 10% на все лекарственные препараты и медицинские изделия для граждан пенсионного возраста и людей с ограниченными возможностями здоровья. Скидка распространяется на все 11 аптек сети в Благовещенске. Круглосуточная служба заказов: +7(914) 555-55-55.',
    NULL,
    NULL,
    'commercial',
    '["pensioners", "disabled"]',
    127.5270,
    50.2670,
    '[28]',
    (SELECT id FROM city WHERE name = 'Благовещенск' LIMIT 1),
    'medicine',
    'Пенсионное удостоверение или справка об инвалидности',
    'Предъявить документ, подтверждающий льготу, в любой аптеке сети при покупке товаров. Скидка предоставляется автоматически на кассе.',
    'https://www.tvoyaapteka.ru/adresa-aptek/',
    '["popular", "recommended"]',
    0,
    UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', '')),
    NOW(),
    NOW(),
    NULL
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- Удаление benefit организации
DELETE FROM benefit WHERE organization_id = UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', ''));

-- Удаление аптек организации
DELETE FROM organization_building WHERE organization_id = UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', ''));

-- Удаление организации
DELETE FROM organization WHERE id = UNHEX(REPLACE('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '-', ''));
