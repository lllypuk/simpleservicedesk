# Coding Standards - Стандарты кодирования Go

## 🎯 Основные принципы

### Go Best Practices
- **gofmt** - стандартное форматирование Go
- **golangci-lint** - комплексная проверка качества
- **goimports** - автоматическая организация импортов
- **Effective Go** - следуем официальным рекомендациям

### Clean Code Principles
- **Single Responsibility** - одна функция = одна ответственность
- **Explicit over Implicit** - явное лучше неявного
- **No Magic Numbers** - используем именованные константы
- **Error Handling** - всегда обрабатываем ошибки

## 📁 File Organization

### Package Structure
```go
// internal/domain/users/user.go
package users

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

// Константы в начале файла
const (
    MaxNameLength = 100
    MinNameLength = 1
)

// Типы и интерфейсы
type Role string

const (
    RoleUser  Role = "User"
    RoleAgent Role = "Agent"
    RoleAdmin Role = "Admin"
)

// Основные структуры
type User struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      Role      `json:"role"`
    CreatedAt time.Time `json:"createdAt"`
}

// Конструкторы и методы
func NewUser(name, email string) (*User, error) {
    // Implementation
}
```

### Import Organization
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. Third-party packages
    "github.com/google/uuid"
    "github.com/labstack/echo/v4"

    // 3. Local packages (с local prefix)
    "simpleservicedesk/internal/domain/users"
    "simpleservicedesk/pkg/logger"
)
```

## 🏗️ Naming Conventions

### Variables and Functions
```go
// Good - описательные имена
func CreateUser(name, email string) (*User, error)
func GetUserByID(id uuid.UUID) (*User, error)

var userRepository UserRepository
var maxRetryAttempts = 3

// Bad - сокращения и неясные имена
func Create(n, e string) (*User, error)
func GetByID(id uuid.UUID) (*User, error)

var repo UserRepository
var max = 3
```

### Types and Interfaces
```go
// Interfaces - существительное с суффиксом -er
type UserRepository interface {
    CreateUser(ctx context.Context, user *User) error
    GetUser(ctx context.Context, id uuid.UUID) (*User, error)
}

// Structs - существительные
type CreateUserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Enums - тип + константы с префиксом
type Status string

const (
    StatusOpen       Status = "Open"
    StatusInProgress Status = "InProgress"
    StatusResolved   Status = "Resolved"
    StatusClosed     Status = "Closed"
)
```

## 🔧 Function Design

### Function Signatures
```go
// Good - явные параметры и возвращаемые значения
func CreateUser(ctx context.Context, name, email, password string) (*User, error) {
    // Implementation
}

// Good - использование struct для множественных параметров
type CreateUserParams struct {
    Name     string
    Email    string
    Password string
    Role     Role
}

func CreateUserWithParams(ctx context.Context, params CreateUserParams) (*User, error) {
    // Implementation
}
```

### Error Handling
```go
// Good - всегда проверяем ошибки
user, err := userRepo.GetUser(ctx, userID)
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// Good - wrap errors с контекстом
if err := userRepo.CreateUser(ctx, user); err != nil {
    return fmt.Errorf("failed to create user %s: %w", user.Email, err)
}

// Bad - игнорирование ошибок
user, _ := userRepo.GetUser(ctx, userID)
```

### Context Usage
```go
// Good - context как первый параметр
func GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    // Передаем context дальше
    return repository.GetUser(ctx, id)
}

// Good - проверяем cancellation
func ProcessUsers(ctx context.Context, users []User) error {
    for _, user := range users {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := processUser(ctx, user); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## 📝 Documentation

### Public API Documentation
```go
// CreateUser creates a new user with the provided name, email and password.
// It validates the input data and returns an error if validation fails.
// The password is hashed before storing in the repository.
//
// Returns:
//   - *User: the created user with generated ID and timestamps
//   - error: validation error, conflict error, or internal error
func CreateUser(ctx context.Context, name, email, password string) (*User, error) {
    // Implementation
}
```

### Internal Comments
```go
func (h *Handler) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return err
    }

    // Hash password before creating user
    hashedPassword, err := hashPassword(req.Password)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user through application layer
    user, err := h.createHandler.CreateUser(c.Request().Context(), req.Name, req.Email, hashedPassword)
    if err != nil {
        return h.handleError(c, err)
    }

    return c.JSON(http.StatusCreated, user)
}
```

## 🧪 Testing Standards

### Test Function Names
```go
// Pattern: Test{FunctionName}_{Scenario}
func TestCreateUser_Success(t *testing.T) {}
func TestCreateUser_ValidationError(t *testing.T) {}
func TestCreateUser_EmailAlreadyExists(t *testing.T) {}

// Table-driven tests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "invalid", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Structure
```go
func TestCreateUser_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewCreateHandler(mockRepo)

    req := CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "password123",
    }

    // Act
    user, err := handler.CreateUser(context.Background(), req)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, req.Name, user.Name)
    assert.Equal(t, req.Email, user.Email)
    assert.NotEmpty(t, user.ID)
    mockRepo.AssertCalled(t, "CreateUser", mock.Anything, mock.Anything)
}
```

## 🔒 Security Guidelines

### Input Validation
```go
func validateUserInput(name, email string) error {
    if strings.TrimSpace(name) == "" {
        return errors.New("name is required")
    }

    if len(name) > MaxNameLength {
        return fmt.Errorf("name must be less than %d characters", MaxNameLength)
    }

    if !isValidEmail(email) {
        return errors.New("invalid email format")
    }

    return nil
}
```

### Password Handling
```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) ([]byte, error) {
    // Никогда не логируем пароли
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func verifyPassword(hashedPassword []byte, password string) error {
    return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}
```

## 📊 Performance Guidelines

### Efficient Code Patterns
```go
// Good - используем string builder для concatenation
var builder strings.Builder
builder.WriteString("Hello")
builder.WriteString(" World")
result := builder.String()

// Good - pre-allocate slices with known capacity
users := make([]User, 0, expectedCount)

// Good - используем context.WithTimeout для внешних вызовов
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

### Memory Management
```go
// Good - закрываем ресурсы
file, err := os.Open("filename")
if err != nil {
    return err
}
defer file.Close()

// Good - используем sync.Pool для частых allocations
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func getUserFromPool() *User {
    return userPool.Get().(*User)
}

func returnUserToPool(user *User) {
    // Reset user fields
    *user = User{}
    userPool.Put(user)
}
```

---

> 💡 **Помни**: Код читается чаще, чем пишется. Приоритизируй читаемость и ясность над краткостью.
