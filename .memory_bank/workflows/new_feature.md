# New Feature Workflow - Процесс добавления новых функций

## 🎯 Когда использовать
- Добавление новой бизнес-функциональности
- Расширение существующих возможностей
- Реализация feature requests
- Внедрение новых API endpoints

## 📋 Пошаговый процесс

### 1. 📖 Планирование и анализ

#### Создание спецификации
```bash
# Создай спецификацию на основе шаблона
cp .memory_bank/specs/feature_xyz.md .memory_bank/specs/your_feature.md

# Заполни все секции:
# - Описание фичи и проблемы
# - Функциональные/нефункциональные требования
# - Техническая архитектура
# - User stories с критериями приемки
# - План тестирования
```

#### Анализ архитектурного воздействия
```bash
# Изучи существующий код
rg "similar_functionality" --type go
ls -la internal/domain/     # Какие domain entities затронуты?
ls -la internal/application/ # Какие use cases нужны?

# Проверь API спецификацию
cat api/openapi.yaml | grep -A 10 -B 5 "relevant_endpoints"
```

### 2. 🏗️ API Design

#### Обновление OpenAPI спецификации
```yaml
# api/openapi.yaml
paths:
  /your-feature:
    post:
      summary: Create new feature entity
      description: Detailed description of what this endpoint does
      tags:
        - your-feature
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateFeatureRequest'
      responses:
        '201':
          description: Feature created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/FeatureResponse'
        '400':
          $ref: '#/components/responses/ValidationError'
        '500':
          $ref: '#/components/responses/InternalError'

components:
  schemas:
    CreateFeatureRequest:
      type: object
      required:
        - name
        - description
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        description:
          type: string
          minLength: 1
          maxLength: 500
        category:
          type: string
          enum: [TypeA, TypeB, TypeC]

    FeatureResponse:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        description:
          type: string
        category:
          type: string
        createdAt:
          type: string
          format: date-time
```

#### Генерация кода из спецификации
```bash
# Генерируй новый код после изменений в API
make generate

# Проверь, что код сгенерировался корректно
ls -la generated/openapi/
git diff generated/ # Просмотри изменения
```

### 3. 🎯 Domain Layer Development

#### Создание domain entities
```go
// internal/domain/yourfeature/feature.go
package yourfeature

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

// Enum для категорий
type Category string

const (
    CategoryTypeA Category = "TypeA"
    CategoryTypeB Category = "TypeB"
    CategoryTypeC Category = "TypeC"
)

// Основная доменная сущность
type Feature struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Category    Category  `json:"category"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// Конструктор с валидацией
func NewFeature(name, description string, category Category) (*Feature, error) {
    if err := validateFeatureData(name, description, category); err != nil {
        return nil, err
    }

    return &Feature{
        ID:          uuid.New(),
        Name:        name,
        Description: description,
        Category:    category,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }, nil
}

// Бизнес-методы
func (f *Feature) UpdateDescription(newDescription string) error {
    if len(strings.TrimSpace(newDescription)) == 0 {
        return errors.New("description cannot be empty")
    }

    f.Description = newDescription
    f.UpdatedAt = time.Now()
    return nil
}

// Приватная валидация
func validateFeatureData(name, description string, category Category) error {
    if len(strings.TrimSpace(name)) == 0 {
        return ValidationError{Field: "name", Message: "name is required"}
    }

    if len(name) > 100 {
        return ValidationError{Field: "name", Message: "name must be less than 100 characters"}
    }

    if len(strings.TrimSpace(description)) == 0 {
        return ValidationError{Field: "description", Message: "description is required"}
    }

    if !isValidCategory(category) {
        return ValidationError{Field: "category", Message: "invalid category"}
    }

    return nil
}

func isValidCategory(category Category) bool {
    validCategories := []Category{CategoryTypeA, CategoryTypeB, CategoryTypeC}
    for _, valid := range validCategories {
        if category == valid {
            return true
        }
    }
    return false
}

// Кастомные ошибки
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

#### Unit tests для domain layer
```go
// internal/domain/yourfeature/feature_test.go
func TestNewFeature_Success(t *testing.T) {
    // Arrange
    name := "Test Feature"
    description := "Test Description"
    category := CategoryTypeA

    // Act
    feature, err := NewFeature(name, description, category)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, name, feature.Name)
    assert.Equal(t, description, feature.Description)
    assert.Equal(t, category, feature.Category)
    assert.NotEmpty(t, feature.ID)
    assert.NotZero(t, feature.CreatedAt)
}

func TestNewFeature_ValidationErrors(t *testing.T) {
    tests := []struct {
        name        string
        featureName string
        description string
        category    Category
        wantError   string
    }{
        {"empty name", "", "desc", CategoryTypeA, "name is required"},
        {"long name", strings.Repeat("a", 101), "desc", CategoryTypeA, "name must be less than 100 characters"},
        {"empty description", "name", "", CategoryTypeA, "description is required"},
        {"invalid category", "name", "desc", "InvalidCategory", "invalid category"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewFeature(tt.featureName, tt.description, tt.category)

            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantError)
        })
    }
}
```

### 4. 🔗 Application Layer Development

#### Repository interface
```go
// internal/application/interfaces.go
type FeatureRepository interface {
    CreateFeature(ctx context.Context, feature *yourfeature.Feature) error
    GetFeature(ctx context.Context, id uuid.UUID) (*yourfeature.Feature, error)
    UpdateFeature(ctx context.Context, id uuid.UUID, updateFn func(*yourfeature.Feature) error) (*yourfeature.Feature, error)
    ListFeatures(ctx context.Context, filters FeatureFilters) ([]*yourfeature.Feature, error)
    DeleteFeature(ctx context.Context, id uuid.UUID) error
}

type FeatureFilters struct {
    Category *yourfeature.Category
    Page     int
    Limit    int
}
```

#### Use case handlers
```go
// internal/application/yourfeature/create.go
package yourfeature

import (
    "context"
    "fmt"

    "simpleservicedesk/internal/domain/yourfeature"
)

type CreateHandler struct {
    featureRepo FeatureRepository
}

func NewCreateHandler(featureRepo FeatureRepository) *CreateHandler {
    return &CreateHandler{featureRepo: featureRepo}
}

type CreateFeatureRequest struct {
    Name        string                    `json:"name"`
    Description string                    `json:"description"`
    Category    yourfeature.Category      `json:"category"`
}

type CreateFeatureError struct {
    Type    string
    Message string
    Cause   error
}

func (e *CreateFeatureError) Error() string {
    return e.Message
}

func (h *CreateHandler) CreateFeature(ctx context.Context, req CreateFeatureRequest) (*yourfeature.Feature, error) {
    // 1. Domain validation через конструктор
    feature, err := yourfeature.NewFeature(req.Name, req.Description, req.Category)
    if err != nil {
        return nil, &CreateFeatureError{
            Type:    "validation",
            Message: "Invalid feature data",
            Cause:   err,
        }
    }

    // 2. Сохранение через repository
    if err := h.featureRepo.CreateFeature(ctx, feature); err != nil {
        return nil, &CreateFeatureError{
            Type:    "internal",
            Message: "Failed to create feature",
            Cause:   err,
        }
    }

    return feature, nil
}
```

#### HTTP handlers
```go
// internal/application/yourfeature/handlers.go
package yourfeature

import (
    "net/http"
    "log/slog"

    "github.com/labstack/echo/v4"
    "simpleservicedesk/generated/openapi"
)

type Handler struct {
    createHandler *CreateHandler
    getHandler    *GetHandler
}

func NewHandler(featureRepo FeatureRepository) *Handler {
    return &Handler{
        createHandler: NewCreateHandler(featureRepo),
        getHandler:    NewGetHandler(featureRepo),
    }
}

// Реализация generated interface
func (h *Handler) PostYourFeature(c echo.Context) error {
    var req openapi.CreateFeatureRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
            Error: openapi.ErrorDetails{
                Code:    "INVALID_REQUEST",
                Message: "Invalid request format",
            },
        })
    }

    // Конвертация из generated types в application types
    createReq := CreateFeatureRequest{
        Name:        req.Name,
        Description: req.Description,
        Category:    yourfeature.Category(req.Category),
    }

    feature, err := h.createHandler.CreateFeature(c.Request().Context(), createReq)
    if err != nil {
        return h.handleError(c, err)
    }

    // Конвертация в response type
    response := openapi.FeatureResponse{
        Id:          feature.ID.String(),
        Name:        feature.Name,
        Description: feature.Description,
        Category:    string(feature.Category),
        CreatedAt:   feature.CreatedAt,
    }

    return c.JSON(http.StatusCreated, response)
}

func (h *Handler) handleError(c echo.Context, err error) error {
    var createErr *CreateFeatureError
    if errors.As(err, &createErr) {
        switch createErr.Type {
        case "validation":
            return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
                Error: openapi.ErrorDetails{
                    Code:    "VALIDATION_ERROR",
                    Message: createErr.Message,
                    Details: extractValidationDetails(createErr.Cause),
                },
            })
        case "internal":
            slog.Error("Internal error creating feature", "error", createErr.Cause)
            return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
                Error: openapi.ErrorDetails{
                    Code:    "INTERNAL_ERROR",
                    Message: "Internal server error",
                },
            })
        }
    }

    slog.Error("Unknown error", "error", err)
    return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
        Error: openapi.ErrorDetails{
            Code:    "INTERNAL_ERROR",
            Message: "Internal server error",
        },
    })
}
```

### 5. 🗄️ Infrastructure Layer Development

#### Repository implementation
```go
// internal/infrastructure/yourfeature/mongo.go
package yourfeature

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/google/uuid"

    "simpleservicedesk/internal/domain/yourfeature"
    "simpleservicedesk/internal/application"
)

type MongoFeatureRepository struct {
    collection *mongo.Collection
}

func NewMongoFeatureRepository(db *mongo.Database) *MongoFeatureRepository {
    return &MongoFeatureRepository{
        collection: db.Collection("features"),
    }
}

// Реализация интерфейса
func (r *MongoFeatureRepository) CreateFeature(ctx context.Context, feature *yourfeature.Feature) error {
    doc := r.featureToDoc(feature)

    _, err := r.collection.InsertOne(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to insert feature: %w", err)
    }

    return nil
}

func (r *MongoFeatureRepository) GetFeature(ctx context.Context, id uuid.UUID) (*yourfeature.Feature, error) {
    var doc featureDoc
    err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, application.ErrFeatureNotFound
        }
        return nil, fmt.Errorf("failed to find feature: %w", err)
    }

    return r.docToFeature(doc), nil
}

// Внутренние типы для MongoDB
type featureDoc struct {
    ID          string    `bson:"_id"`
    Name        string    `bson:"name"`
    Description string    `bson:"description"`
    Category    string    `bson:"category"`
    CreatedAt   time.Time `bson:"createdAt"`
    UpdatedAt   time.Time `bson:"updatedAt"`
}

func (r *MongoFeatureRepository) featureToDoc(f *yourfeature.Feature) featureDoc {
    return featureDoc{
        ID:          f.ID.String(),
        Name:        f.Name,
        Description: f.Description,
        Category:    string(f.Category),
        CreatedAt:   f.CreatedAt,
        UpdatedAt:   f.UpdatedAt,
    }
}

func (r *MongoFeatureRepository) docToFeature(doc featureDoc) *yourfeature.Feature {
    id, _ := uuid.Parse(doc.ID) // В production добавь error handling

    return &yourfeature.Feature{
        ID:          id,
        Name:        doc.Name,
        Description: doc.Description,
        Category:    yourfeature.Category(doc.Category),
        CreatedAt:   doc.CreatedAt,
        UpdatedAt:   doc.UpdatedAt,
    }
}
```

### 6. 🧪 Тестирование

#### Integration tests
```go
// integration_test/yourfeature/api_test.go
func TestFeatureAPI_CreateAndGet(t *testing.T) {
    suite := setupTestSuite(t)
    defer suite.Cleanup()

    // Test data
    createReq := map[string]interface{}{
        "name":        "Test Feature",
        "description": "Test Description",
        "category":    "TypeA",
    }

    // Create feature
    resp, err := suite.Client.Post("/your-feature", createReq)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createResp openapi.FeatureResponse
    err = json.NewDecoder(resp.Body).Decode(&createResp)
    require.NoError(t, err)

    // Validate created feature
    assert.Equal(t, createReq["name"], createResp.Name)
    assert.Equal(t, createReq["description"], createResp.Description)
    assert.NotEmpty(t, createResp.Id)

    // Get created feature
    resp, err = suite.Client.Get(fmt.Sprintf("/your-feature/%s", createResp.Id))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### 7. 🔧 Интеграция в приложение

#### Регистрация handlers в main
```go
// cmd/server/main.go или internal/run.go
func setupRoutes(e *echo.Echo, app *application.Suite) {
    // Existing handlers...

    // New feature handlers
    featureHandler := yourfeature.NewHandler(app.FeatureRepo)

    // Register routes (matches OpenAPI paths)
    e.POST("/your-feature", featureHandler.PostYourFeature)
    e.GET("/your-feature/:id", featureHandler.GetYourFeatureID)
}
```

#### Обновление application suite
```go
// internal/application/suite.go
type Suite struct {
    UserRepo    UserRepository
    FeatureRepo FeatureRepository // Добавь новый repository
}

func New(userRepo UserRepository, featureRepo FeatureRepository) *Suite {
    return &Suite{
        UserRepo:    userRepo,
        FeatureRepo: featureRepo,
    }
}
```

### 8. 📝 Финализация

#### Проверка всей системы
```bash
# Генерация кода
make generate

# Линтинг
make lint

# Полное тестирование
make test

# Проверка покрытия
make coverage_report
```

#### Git workflow
```bash
# Создай feature branch
git checkout -b feature/your-feature-name

# Commit по частям
git add internal/domain/yourfeature/
git commit -m "feat: add yourfeature domain entities with validation"

git add internal/application/yourfeature/
git commit -m "feat: add yourfeature application layer with use cases"

git add internal/infrastructure/yourfeature/
git commit -m "feat: add yourfeature MongoDB repository implementation"

git add api/openapi.yaml generated/
git commit -m "feat: add yourfeature API endpoints and generated code"

git add integration_test/
git commit -m "feat: add yourfeature integration tests"

# Final commit
git add .
git commit -m "feat: integrate yourfeature into main application

Complete implementation of YourFeature functionality including:
- Domain entities with business logic validation
- Application layer with use cases and error handling
- MongoDB repository implementation
- REST API endpoints with OpenAPI specification
- Comprehensive test coverage (unit + integration)

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

## ⚠️ Checklist для завершения

### Перед созданием PR
- [ ] Все тесты проходят (`make test`)
- [ ] Код проходит линтинг (`make lint`)
- [ ] API документация обновлена
- [ ] Generated code актуален (`make generate`)
- [ ] Integration tests покрывают основные сценарии
- [ ] Error handling следует проектным паттернам
- [ ] Логирование добавлено для важных операций
- [ ] Performance не пострадал

### Code review checklist
- [ ] Соответствие Clean Architecture принципам
- [ ] Правильное разделение ответственности между слоями
- [ ] Comprehensive error handling
- [ ] Security considerations (input validation, auth)
- [ ] Database indexes добавлены если нужно
- [ ] API backward compatibility сохранена

---

> 💡 **Принцип**: Разрабатывай feature инкрементально - domain first, затем application, infrastructure, и наконец API integration. Каждый слой должен быть независимо тестируемым.
