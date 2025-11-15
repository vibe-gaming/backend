-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- Добавляем FULLTEXT индекс для поиска по title и description
ALTER TABLE benefit 
ADD FULLTEXT INDEX ft_benefit_search (title, description);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- Удаляем FULLTEXT индекс
ALTER TABLE benefit 
DROP INDEX ft_benefit_search;
