-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- Вставка организации "Твоя Аптека" для Якутска
INSERT INTO organization (id, name, description, created_at, updated_at, deleted_at)
VALUES (
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'Твоя Аптека',
    'Аптечная сеть в Якутске. Круглосуточная аптечная служба заказов: +7(914) 555-55-55. Широкий ассортимент лекарств, медицинских изделий, косметики и товаров для здоровья.',
    NOW(),
    NOW(),
    NULL
);

-- Вставка аптек организации в Якутске
INSERT INTO organization_building (id, organization_id, address, latitude, longitude, phone_number, start_time, end_time, is_open, tags, created_at, updated_at, deleted_at)
VALUES
-- Аптека 1: Ул. Октябрьская, 1
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Октябрьская, 1',
    62.0355, 129.7422,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 2: Ул. Аммосова, 8
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Аммосова, 8',
    62.0275, 129.7315,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 22:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 3: Ул. Кирова, 18
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Кирова, 18',
    62.0312, 129.7288,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 4: Ул. Ленина, 24
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Ленина, 24',
    62.0298, 129.7352,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 5: Ул. Орджоникидзе, 47
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Орджоникидзе, 47',
    62.0335, 129.7410,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 6: Проспект Ленина, 1 (ТЦ Якутск)
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, проспект Ленина, 1 (ТЦ Якутск)',
    62.0285, 129.7365,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 22:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 7: Ул. Дзержинского, 16
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Дзержинского, 16',
    62.0265, 129.7245,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 8: Ул. Курашова, 28
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Курашова, 28',
    62.0395, 129.7485,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 9: Ул. Петровского, 12/1
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Петровского, 12/1',
    62.0245, 129.7198,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 10: Ул. Хабарова, 15
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Хабарова, 15',
    62.0318, 129.7445,
    '+79145555555',
    '2025-01-01 09:00:00',
    '2025-01-01 21:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 11: Ул. Автодорожная, 50
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Автодорожная, 50',
    62.0425, 129.7520,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 22:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
),
-- Аптека 12: Ул. Богдана Чижика, 8
(
    UNHEX(REPLACE(UUID(), '-', '')),
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    'г. Якутск, ул. Богдана Чижика, 8',
    62.0365, 129.7395,
    '+79145555555',
    '2025-01-01 08:00:00',
    '2025-01-01 20:00:00',
    TRUE,
    '[]',
    NOW(), NOW(), NULL
);

-- Вставка benefit для аптечной сети "Твоя Аптека" в Якутске
INSERT INTO benefit (
    id, title, description, valid_from, valid_to, type,
    target_group_ids, longitude, latitude, region, city_id,
    category, requirment, how_to_use, source_url, tags, views,
    organization_id, created_at, updated_at, deleted_at
)
VALUES (
    UNHEX(REPLACE(UUID(), '-', '')),
    'Скидка 10% для пенсионеров и инвалидов в сети аптек "Твоя Аптека" (Якутск)',
    'Аптечная сеть "Твоя Аптека" предоставляет постоянную скидку 10% на все лекарственные препараты и медицинские изделия для граждан пенсионного возраста и людей с ограниченными возможностями здоровья. Скидка распространяется на все 12 аптек сети в Якутске. Круглосуточная служба заказов: +7(914) 555-55-55.',
    NULL,
    NULL,
    'commercial',
    '["pensioners", "disabled"]',
    129.7422,
    62.0355,
    '[14]',
    (SELECT id FROM city WHERE name = 'Якутск' LIMIT 1),
    'medicine',
    'Пенсионное удостоверение или справка об инвалидности',
    'Предъявить документ, подтверждающий льготу, в любой аптеке сети при покупке товаров. Скидка предоставляется автоматически на кассе.',
    'https://www.tvoyaapteka.ru/adresa-aptek/',
    '["popular", "recommended"]',
    0,
    UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', '')),
    NOW(),
    NOW(),
    NULL
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- Удаление benefit организации
DELETE FROM benefit WHERE organization_id = UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', ''));

-- Удаление аптек организации
DELETE FROM organization_building WHERE organization_id = UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', ''));

-- Удаление организации
DELETE FROM organization WHERE id = UNHEX(REPLACE('b2c3d4e5-f6a7-8901-bcde-f12345678901', '-', ''));

