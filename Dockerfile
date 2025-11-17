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

# Создаём папку для шрифтов и скачиваем DejaVu Sans напрямую из SourceForge
RUN mkdir -p /app/fonts && \
    curl -L -o /tmp/dejavu-fonts.tar.bz2 "https://downloads.sourceforge.net/project/dejavu/dejavu/2.37/dejavu-fonts-ttf-2.37.tar.bz2" && \
    cd /tmp && tar -xjf dejavu-fonts.tar.bz2 && \
    cp dejavu-fonts-ttf-2.37/ttf/DejaVuSans.ttf /app/fonts/ && \
    rm -rf /tmp/dejavu-fonts* && \
    chmod 644 /app/fonts/DejaVuSans.ttf && \
    echo "📦 Downloaded font info:" && \
    ls -lh /app/fonts/DejaVuSans.ttf && \
    echo "🔍 First 16 bytes (magic number):" && \
    head -c 16 /app/fonts/DejaVuSans.ttf | od -A n -t x1 && \
    echo "✅ Font downloaded and extracted successfully"

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app && \
    echo "👤 User created and permissions set:" && \
    ls -la /app/

USER appuser

# Проверяем доступность шрифта из-под appuser
RUN echo "🔍 Checking font access as appuser:" && \
    ls -la /app/fonts/ && \
    test -r /app/fonts/DejaVuSans.ttf && \
    echo "✅ Font is readable by appuser"

EXPOSE 8080

CMD ["./main"]

