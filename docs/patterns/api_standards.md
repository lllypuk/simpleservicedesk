# API Standards - Стандарты проектирования API

## 🎯 Общие принципы

### RESTful Design
- **Resource-based URLs** - `/users/{id}`, `/tickets/{id}`
- **HTTP методы** - GET (чтение), POST (создание), PUT (обновление), DELETE (удаление)
- **Статус коды** - используем стандартные HTTP статусы
- **JSON** - единственный формат данных

### OpenAPI First
- **Спецификация** - `api/openapi.yaml` как источник истины
- **Code Generation** - server и client код генерируется автоматически
- **Validation** - автоматическая валидация через generated code

## 📋 Naming Conventions

### URL Patterns
```
GET    /users              # Список пользователей
GET    /users/{id}         # Конкретный пользователь
POST   /users              # Создание пользователя
PUT    /users/{id}         # Обновление пользователя
DELETE /users/{id}         # Удаление пользователя
```

### Response Structure
```json
// Успешный ответ с данными
{
  "id": "uuid",
  "name": "string",
  "email": "string",
  "role": "User|Agent|Admin",
  "createdAt": "2024-01-01T00:00:00Z"
}

// Список ресурсов
{
  "users": [...],
  "total": 100,
  "page": 1,
  "limit": 20
}

// Ошибка
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": {...}
  }
}
```

## 🔧 Technical Standards

### UUID Identifiers
- Все ID используют UUID v4
- Генерация через `github.com/google/uuid`
- В URL path parameters как строки

### Timestamps
- **RFC3339 format** - `2024-01-01T00:00:00Z`
- **UTC timezone** - всегда UTC в API
- **Naming** - `createdAt`, `updatedAt`, `deletedAt`

### Validation
- **Required fields** - обязательные поля в schema
- **Format validation** - email, UUID через OpenAPI
- **Business validation** - в domain layer

## 🚦 HTTP Status Codes

### Success (2xx)
- `200 OK` - успешное получение/обновление
- `201 Created` - успешное создание
- `204 No Content` - успешное удаление

### Client Errors (4xx)
- `400 Bad Request` - невалидные данные
- `401 Unauthorized` - отсутствует или невалидный Bearer токен
- `403 Forbidden` - недостаточно прав для выполнения операции
- `404 Not Found` - ресурс не найден
- `409 Conflict` - конфликт данных
- `429 Too Many Requests` - превышен rate limit; ответ включает заголовок `Retry-After`

### Server Errors (5xx)
- `500 Internal Server Error` - внутренняя ошибка

## 🔒 Security Standards

### Authentication
- **Bearer Token** — передается в заголовке `Authorization: Bearer <jwt>`
- **JWT structure** — claims: `user_id` (UUID), `role` (admin/agent/customer), standard JWT fields
- **Login endpoint** — `POST /login` (public); все остальные эндпоинты требуют валидный токен
- **No self-registration** — пользователей создают только Admin

### Authorization
- **Role-based Access** — проверка ролей на уровне middleware (не handlers)
- `admin` — полный доступ ко всем ресурсам
- `agent` — список пользователей, управление тикетами
- `customer` — только собственные тикеты и комментарии
- Owner-or-admin pattern для `GET/PUT /users/{id}`

### Rate Limiting
- Глобальный лимит: `RATE_LIMIT_RPS` (default 100 req/s)
- Stricter limit на `POST /login`: 5 req/min per client
- При превышении: `429 Too Many Requests` + `Retry-After` header

### Input Validation
- **Schema validation** - через OpenAPI generated code (middleware)
- **SQL/NoSQL Injection** - защита через параметризованные запросы MongoDB

## 📝 Documentation Standards

### OpenAPI Schema
```yaml
components:
  schemas:
    User:
      type: object
      required:
        - name
        - email
      properties:
        id:
          type: string
          format: uuid
          readOnly: true
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
```

### Response Examples
```yaml
responses:
  '200':
    description: User retrieved successfully
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/User'
        example:
          id: "550e8400-e29b-41d4-a716-446655440000"
          name: "John Doe"
          email: "john@example.com"
```

## 🔄 Versioning Strategy

### Current Approach
- **No versioning** - пока API стабильный
- **Breaking changes** - требуют major version bump
- **Backward compatibility** - приоритет при изменениях

### Future Versioning
```
/v1/users/{id}    # Когда потребуется версионирование
/v2/users/{id}    # Новая версия API
```

## ⚡ Performance Guidelines

### Pagination
```yaml
parameters:
  - name: page
    in: query
    schema:
      type: integer
      minimum: 1
      default: 1
  - name: limit
    in: query
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 20
```

### Filtering
```yaml
parameters:
  - name: role
    in: query
    schema:
      type: string
      enum: [User, Agent, Admin]
```

---

> 💡 **Принцип**: API должен быть предсказуемым и следовать REST conventions. OpenAPI спецификация - единственный источник истины для API контракта.
