# Simple ServiceDesk

- [![Go Reference](https://pkg.go.dev/badge/simpleservicedesk)](https://pkg.go.dev/simpleservicedesk) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Простой веб-сервис на Go для управления пользователями через RESTful API.

## Содержание
- [Обзор](#обзор)
- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Начало работы](#начало-работы)
  - [Требования](#требования)
  - [Установка](#установка)
  - [Настройка](#настройка)
  - [Генерация кода](#генерация-кода)
  - [Запуск](#запуск)
  - [Контейнеризация](#контейнеризация)
- [Примеры использования](#примеры-использования)
- [Документация API](#документация-api)
- [Тестирование](#тестирование)
- [Contributing](#contributing)
- [Лицензия](#лицензия)

## Обзор
Simple ServiceDesk — минималистичный веб-сервис на Go, предоставляющий базовые API для управления пользователями.

## Возможности
- Создание и получение пользователей через RESTful API
- Спецификация OpenAPI 3.0 и генерация кода (сервер и клиент)
- In-memory репозиторий для быстрой разработки и тестирования
- Структурированное логирование с slog
- Корректное завершение работы (graceful shutdown) и обработка сигналов

## Архитектура
- **cmd/server**: точка входа приложения
- **internal/application**: настройка HTTP-сервера и маршрутов
- **internal/infrastructure/users**: репозиторий пользователей в памяти
- **internal/domain/users**: бизнес-логика и сущности пользователей
- **pkg**: общие утилиты и middleware
- **api/openapi.yaml**: спецификация OpenAPI
- **generated**: сгенерированные типы, сервер и клиент из спецификации OpenAPI

## Начало работы

### Требования
- Go версии 1.20 или выше
- oapi-codegen (установите с помощью `go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest`)

### Установка
```bash
git clone https://github.com/<YOUR_USERNAME>/simpleservicedesk.git
cd simpleservicedesk
```

### Настройка
Установите необходимые переменные окружения (значения по умолчанию указаны в скобках):
```bash
ENV_TYPE (default: Testing)
SERVER_PORT (default: 8080)
INTERRUPT_TIMEOUT (default: 2s)
READ_HEADER_TIMEOUT (default: 5s)
```

### Генерация кода
Сгенерируйте код по спецификации OpenAPI (типы, заглушки сервера, клиент):
```bash
go generate ./generated
```

### Запуск
Соберите и запустите сервер:
```bash
go build -o server ./cmd/server
./server
# или
go run cmd/server/main.go
```
Сервер будет доступен по адресу `http://localhost:SERVER_PORT`.

### Контейнеризация

Сборка и запуск сервиса в Docker:
```bash
docker build -t simpleservicedesk -f build/Dockerfile .
docker run -d --name simpleservicedesk -p 8080:8080 simpleservicedesk
```

## Примеры использования

Создать пользователя:
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john.doe@example.com"}'
```

Получить пользователя по ID:
```bash
curl http://localhost:8080/users/{id}
```

## Документация API
API описано в файле `api/openapi.yaml`:
- **POST /users**: Создать нового пользователя
- **GET /users/{id}**: Получить данные пользователя по ID

Можно использовать инструменты вроде Swagger UI или генерировать клиент на основе спецификации OpenAPI.

## Тестирование
Запустите модульные и интеграционные тесты:
```bash
go test ./...
```

## Contributing

PRs и issue приветствуются! Пожалуйста, обсудите значительные изменения через issue перед отправкой PR.

## Лицензия
Проект лицензирован на условиях MIT License. Подробнее см. в файле [LICENSE](LICENSE).
