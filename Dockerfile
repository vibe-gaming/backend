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

# Устанавливаем CA certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарник из builder stage
COPY --from=builder /build/main .

# Копируем шрифты для генерации PDF из builder stage
COPY --from=builder /build/fonts /app/fonts

# Проверяем, что шрифты скопировались (для отладки)
RUN ls -la /app/fonts/ || echo "Fonts directory not found!"

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./main"]

