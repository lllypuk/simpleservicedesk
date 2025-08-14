# Tech Stack - Технологический паспорт

## 🏗️ Архитектурный подход

**Clean Architecture** - четкое разделение слоев для поддерживаемости и тестируемости:
- Domain Layer - бизнес-логика и сущности
- Application Layer - use cases и интерфейсы
- Infrastructure Layer - внешние зависимости

## 🔧 Основной стек

### Backend
- **Go 1.24** - основной язык разработки
- **Echo v4** - HTTP router и middleware
- **MongoDB** - основная база данных
- **In-Memory Storage** - fallback для тестирования

### API & Code Generation
- **OpenAPI 3.0** - спецификация API (`api/openapi.yaml`)
- **oapi-codegen** - генерация server/client кода из OpenAPI
- **UUID** - идентификаторы сущностей

### Security & Auth
- **bcrypt** - хеширование паролей
- **Role-based Access Control** - Admin/Agent/User

## 🧪 Тестирование

### Test Stack
- **testify** - assertions и test suites
- **testcontainers-go** - интеграционные тесты с реальной MongoDB
- **Go testing** - встроенный тест-фреймворк

### Test Strategy
- **Unit tests** - `internal/...` (бизнес-логика)
- **Integration tests** - `integration_test/...` (с БД)
- **Coverage reporting** - HTML отчеты

## 🛠️ Development Tools

### Code Quality
- **golangci-lint** - линтинг
- **goimports** - форматирование импортов
- **go fmt** - форматирование кода

### Performance
- **pprof** - CPU и memory profiling
- **net/http/pprof** - веб-интерфейс для профилирования

## 📦 Deployment & Infrastructure

### Containerization
- **Docker** - контейнеризация приложения
- **Docker Compose** - локальная разработка с MongoDB

### Environment
- **Environment variables** - конфигурация
- **.env files** - локальная разработка

## 📚 Key Libraries

```go
// Core dependencies
github.com/labstack/echo/v4           // HTTP framework
go.mongodb.org/mongo-driver           // MongoDB driver
github.com/google/uuid                // UUID generation
golang.org/x/crypto                   // Password hashing

// Code generation
github.com/getkin/kin-openapi         // OpenAPI support
github.com/oapi-codegen/runtime       // Generated code runtime

// Testing
github.com/stretchr/testify           // Test assertions
github.com/testcontainers/testcontainers-go // Container testing

// Concurrency
golang.org/x/sync                     // Extended sync primitives
```

## 🗂️ Project Structure

```
simpleservicedesk/
├── api/                    # OpenAPI specifications
├── cmd/server/             # Application entry point
├── generated/              # Auto-generated code from OpenAPI
├── internal/               # Private application code
│   ├── application/        # Use cases, HTTP handlers
│   ├── domain/            # Business entities and logic
│   └── infrastructure/    # Database repositories
├── integration_test/       # Integration tests
└── pkg/                   # Public packages (middleware, etc.)
```

## ⚙️ Configuration

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | `development` |
| `HTTP_SERVER_PORT` | Server port | `8080` |
| `MONGO_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | Database name | `servicedesk` |

## 🔄 Code Generation Workflow

1. **API Design** - изменения в `api/openapi.yaml`
2. **Generate** - `make generate` создает Go код
3. **Implement** - реализация бизнес-логики
4. **Test** - unit и integration тесты
5. **Lint** - проверка качества кода

## 🚀 Performance Characteristics

### Benchmarks
- **API Response Time** - target < 200ms
- **Concurrent Users** - designed for 1000+ concurrent connections
- **Database** - MongoDB with indexes for performance

### Scalability
- **Stateless design** - horizontal scaling ready
- **Connection pooling** - efficient database usage
- **Clean shutdown** - graceful termination

---

> 💡 **Принцип**: Выбираем проверенные технологии с минимальными внешними зависимостями для долгосрочной поддерживаемости.
