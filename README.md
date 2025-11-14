# Backend для Hack The Ice 7.0

### Инструкция по локальному развертыванию приложения:
1. Поднять БД с помощью - `make compose`
2. Установка goose для накатывания миграций - `make install-goose` 
3. Установка миграций - `make migration-up`
4. Запуск приложения - `make run`

### Дополнительно
1. Rollback миграций - `make migration-down`
2. Статус миграций - `make migration-status`
3. Генерация сваггер документации - `make swag`
4. Запуск линтера - `make lint` 
