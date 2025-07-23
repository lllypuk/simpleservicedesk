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
- Репозитории для In-memory и MongoDB
- Структурированное логирование с slog
- Корректное завершение работы (graceful shutdown) и обработка сигналов
- Контейнеризация с помощью Docker и Docker Compose

## Архитектура
Проект следует принципам чистой архитектуры и разделен на следующие слои:
- **Domain**: Содержит основные бизнес-сущности и логику.
- **Application**: Координирует бизнес-логику и сценарии использования.
- **Infrastructure**: Реализует внешние зависимости, такие как базы данных, API-клиенты и т.д.
- **Generated**: Содержит код, сгенерированный на основе OpenAPI спецификации.

## Начало работы

### Требования
- [Go](https://golang.org/dl/) (версия 1.21 или выше)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

### Установка
Клонируйте репозиторий и установите зависимости:
```bash
git clone https://github.com/your-username/simpleservicedesk.git
cd simpleservicedesk
go mod download
```

### Настройка
Проект настраивается с помощью переменных окружения. Создайте файл `.env` в корне проекта по примеру `.env.example` (если он есть) или используйте переменные напрямую:
```bash
# .env
APP_ENV=development
HTTP_SERVER_PORT=8080
MONGO_URI=mongodb://user:password@localhost:27017
```

### Генерация кода
Для генерации кода сервера и клиента из OpenAPI спецификации выполните:
```bash
make generate
```

### Запуск
#### Локально
Для запуска сервера локально:
```bash
make run
```
Сервер будет доступен по адресу `http://localhost:8080`.

### Контейнеризация
Для запуска сервиса и базы данных MongoDB в Docker-контейнерах:
```bash
docker-compose up -d
```

## Примеры использования
Примеры запросов к API с использованием `curl`.

### Создание пользователя
```bash
curl -X POST http://localhost:8080/users \
-H "Content-Type: application/json" \
-d '{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "password": "securepassword123"
}'
```

### Получение пользователя по ID
Замените `{userId}` на реальный ID пользователя.
```bash
curl -X GET http://localhost:8080/users/{userId}
```

## Документация API
API документировано с использованием спецификации OpenAPI 3.0. Файл спецификации находится здесь: `api/openapi.yaml`.

## Тестирование
Для запуска всех тестов (модульных и интеграционных) выполните:
```bash
make test
```

## Contributing

PRs и issue приветствуются! Пожалуйста, обсудите значительные изменения через issue перед отправкой PR.

## Лицензия
Проект лицензирован на условиях MIT License. Подробнее см. в файле [LICENSE](LICENSE).
