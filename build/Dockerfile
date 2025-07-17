# Многоступенчатая сборка для оптимизации размера образа
FROM golang:1.24-alpine AS builder

# Установка необходимых пакетов для сборки
RUN apk add --no-cache git ca-certificates

# Создание рабочей директории
WORKDIR /app

# Копирование файлов модулей для кэширования зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Финальный минимальный образ
FROM alpine:latest

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

# Создание непривилегированного пользователя
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Создание рабочей директории
WORKDIR /app

# Копирование скомпилированного бинарника из builder образа
COPY --from=builder /app/server .

# Смена владельца файлов
RUN chown -R appuser:appgroup /app

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта (по умолчанию для веб-сервера)
EXPOSE 8080

# Запуск приложения
CMD ["./server"]
