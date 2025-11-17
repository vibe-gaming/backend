# Build stage
FROM golang:1.24.10-alpine AS builder

WORKDIR /build

# Копируем go mod и sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Устанавливаем CA certificates для HTTPS запросов и curl
RUN apk --no-cache add ca-certificates tzdata curl

# Копируем бинарник из builder stage
COPY --from=builder /build/main .

# Создаём папку для шрифтов и скачиваем DejaVu Sans
RUN mkdir -p /app/fonts && \
    curl -L -o /app/fonts/DejaVuSans.ttf "https://github.com/dejavu-fonts/dejavu-fonts/raw/master/ttf/DejaVuSans.ttf" && \
    ls -la /app/fonts/ && \
    echo "✅ Font downloaded successfully"

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./main"]

