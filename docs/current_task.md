# Задача: Добавление API тестов для Categories (интеграционное тестирование)

## Обзор проекта

SimpleServiceDesk - это Go-приложение сервисной службы, построенное по принципам чистой архитектуры с использованием:
- **Web Framework**: Echo v4
- **База данных**: MongoDB с testcontainers для тестирования
- **Кодогенерация**: oapi-codegen из OpenAPI 3.0 спецификации
- **Тестирование**: testcontainers-go для интеграционных тестов с реальным MongoDB

## Анализ текущего состояния

### Структура тестов
Проект использует централизованную структуру интеграционных тестов:
```
test/integration/
├── api/           # HTTP API интеграционные тесты
│   ├── users_test.go      ✅ РЕАЛИЗОВАН
│   ├── tickets_test.go    ✅ РЕАЛИЗОВАН
│   ├── organizations_test.go ✅ РЕАЛИЗОВАН
│   └── categories_test.go ❌ ОТСУТСТВУЕТ
├── repositories/  # Тесты репозиториев с testcontainers
└── shared/       # Общие утилиты и настройки тестов
```

### Существующая реализация Categories API

**OpenAPI Endpoints** (из `api/openapi.yaml`):
- `POST /categories` - Создание категории
- `GET /categories` - Список категорий с иерархическим древом
- `GET /categories/{id}` - Получение категории по ID
- `PUT /categories/{id}` - Обновление категории
- `DELETE /categories/{id}` - Удаление категории
- `GET /categories/{id}/tickets` - Получение билетов категории

**Ключевые особенности Categories API**:
- Иерархическая структура (parent_id, children)
- Фильтрация по organization_id, parent_id, is_active
- Поддержка include_children и include_subcategories
- Валидация существования parent_id и organization_id

**Application Layer** (полностью реализован):
- `internal/application/categories/` - все CRUD операции с unit тестами
- Handlers: create.go, get.go, list.go, update.go, delete.go, tickets.go

**Infrastructure Layer**:
- `internal/infrastructure/categories/mongo.go` - MongoDB репозиторий

## Задача

**Цель**: Создать полное интеграционное тестирование Categories API по образцу существующих тестов users_test.go, tickets_test.go, organizations_test.go.

### Требования к реализации

1. **Файл**: `test/integration/api/categories_test.go`

2. **Структура теста**:
   ```go
   //go:build integration
   // +build integration

   package api_test

   type CategoryAPITestSuite struct {
       shared.IntegrationSuite
   }
   ```

3. **Обязательные тестовые сценарии**:

   **POST /categories** - Создание категории:
   - ✅ Успешное создание root категории
   - ✅ Успешное создание дочерней категории
   - ❌ Невалидные данные (пустое имя, невалидный organization_id)
   - ❌ Несуществующий parent_id
   - ❌ Дублирование имени в organization
   - ❌ Несуществующий organization_id

   **GET /categories** - Список категорий:
   - ✅ Получение всех категорий
   - ✅ Фильтрация по organization_id
   - ✅ Фильтрация по parent_id (root категории)
   - ✅ Фильтрация по is_active
   - ✅ include_children=true/false
   - ✅ Пустой результат для несуществующей organization

   **GET /categories/{id}** - Получение по ID:
   - ✅ Успешное получение существующей категории
   - ❌ Несуществующий ID
   - ❌ Невалидный UUID format

   **PUT /categories/{id}** - Обновление:
   - ✅ Успешное обновление name, description, is_active
   - ✅ Изменение parent_id (перемещение в иерархии)
   - ❌ Невалидные данные
   - ❌ Несуществующий ID
   - ❌ Несуществующий parent_id
   - ❌ Циклические ссылки (parent_id указывает на самого себя или потомка)

   **DELETE /categories/{id}** - Удаление:
   - ✅ Успешное удаление категории без детей
   - ✅ Удаление с каскадным обновлением дочерних категорий
   - ❌ Несуществующий ID
   - ❌ Невалидный UUID

   **GET /categories/{id}/tickets** - Билеты категории:
   - ✅ Получение билетов категории
   - ✅ include_subcategories=true/false
   - ✅ Пустой результат для категории без билетов
   - ❌ Несуществующая категория

4. **Тестовые данные** (добавить в `shared/fixtures.go`):
   ```go
   // TestCategoryData для различных тестовых сценариев
   type TestCategoryData struct {
       Name           string
       Description    string
       OrganizationID uuid.UUID
       ParentID       *uuid.UUID
       IsActive       bool
   }

   var (
       TestCategoryRoot1 = TestCategoryData{...}
       TestCategoryChild1 = TestCategoryData{...}
       // и т.д.
   )
   ```

5. **Соблюдение паттернов**:
   - Использовать `shared.IntegrationSuite` как базу
   - Следовать структуре существующих API тестов
   - Использовать testcontainers для реального MongoDB
   - Тесты должны быть изолированными (SetupTest/TearDownTest)
   - Build tag `//go:build integration`

6. **Валидация ответов**:
   - Проверка HTTP статус кодов
   - Валидация JSON структуры ответов
   - Проверка бизнес-логики (иерархия, фильтры)
   - Проверка error responses

## Команды для разработки

После реализации обязательно запустить:
```bash
make lint                    # Форматирование и линтеры
make test-integration        # Все интеграционные тесты
make test-api               # Только HTTP API тесты
```

## Ожидаемый результат

Файл `test/integration/api/categories_test.go` с полным покрытием всех Categories API endpoints, соответствующий качеству и структуре существующих интеграционных тестов в проекте.

Тесты должны проходить команду `make test-api` и обеспечивать надежную валидацию всей функциональности Categories API.
