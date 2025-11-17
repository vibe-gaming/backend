# Шрифты для PDF генератора

Эта директория содержит TTF шрифты для генерации PDF с поддержкой кириллицы.

## Автоматическая установка

**При сборке Docker образа** шрифт автоматически скачивается из официального репозитория:

```dockerfile
RUN curl -L -o /app/fonts/DejaVuSans.ttf "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans.ttf"
```

Источник: https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans.ttf

## Установленные шрифты

- **DejaVuSans.ttf** - основной шрифт с полной поддержкой кириллицы

## Скачивание шрифта для локальной разработки

Для локальной разработки вне Docker:

```bash
cd fonts
curl -L -o DejaVuSans.ttf "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans.ttf"
```

## Альтернативные источники шрифтов

1. **Системные шрифты Linux:**
   ```
   /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf
   ```

2. **Системные шрифты macOS:**
   ```
   /System/Library/Fonts/Supplemental/Arial Unicode.ttf
   ```

3. **Google Fonts:**
   - Roboto
   - Open Sans
   - Noto Sans

## Информация о DejaVuSans

- **Лицензия**: Free (можно использовать коммерчески)
- **Поддержка**: Latin, Cyrillic, Greek, и многие другие
- **Размер**: ~285 KB
- **Сайт**: https://dejavu-fonts.github.io/

