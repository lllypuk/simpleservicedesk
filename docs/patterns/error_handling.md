# Error Handling - Паттерны обработки ошибок

## 🎯 Общие принципы

### Layered Error Handling
- **Domain errors** - бизнес-логические ошибки в domain layer
- **Application errors** - ошибки координации в application layer
- **Infrastructure errors** - ошибки внешних зависимостей
- **HTTP errors** - преобразование в HTTP responses

### Error Categories
1. **Validation Errors** - невалидные входные данные → `400 Bad Request`
2. **Authentication Errors** - отсутствует или невалидный JWT токен → `401 Unauthorized`
3. **Authorization Errors** - недостаточно прав доступа → `403 Forbidden`
4. **Business Logic Errors** - нарушение бизнес-правил → `409 Conflict` / `404 Not Found`
5. **Rate Limit Errors** - превышен лимит запросов → `429 Too Many Requests`
6. **Infrastructure Errors** - проблемы с БД, сетью и т.д. → `500 Internal Server Error`

## 🏗️ Architecture Patterns

### Domain Layer Errors
```go
// internal/domain/users/user.go
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// Конструктор пользователя с валидацией
func NewUser(name, email string) (*User, error) {
    if name == "" {
        return nil, ValidationError{Field: "name", Message: "name is required"}
    }
    if !isValidEmail(email) {
        return nil, ValidationError{Field: "email", Message: "invalid email format"}
    }
    return &User{Name: name, Email: email}, nil
}
```

### Application Layer Error Handling
```go
// internal/application/users/create.go
type CreateUserError struct {
    Type    string // "validation", "conflict", "internal"
    Message string
    Cause   error
}

func (c *CreateHandler) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // 1. Domain validation
    user, err := users.NewUser(req.Name, req.Email)
    if err != nil {
        return nil, &CreateUserError{
            Type:    "validation",
            Message: "Invalid user data",
            Cause:   err,
        }
    }

    // 2. Repository call with error handling
    createdUser, err := c.userRepo.CreateUser(ctx, req.Email, hashedPassword, func() (*users.User, error) {
        return user, nil
    })
    if err != nil {
        if isDuplicateEmailError(err) {
            return nil, &CreateUserError{
                Type:    "conflict",
                Message: "User with this email already exists",
                Cause:   err,
            }
        }
        return nil, &CreateUserError{
            Type:    "internal",
            Message: "Failed to create user",
            Cause:   err,
        }
    }

    return createdUser, nil
}
```

### Auth Error Patterns

Authentication (401) and authorization (403) errors are handled by middleware before reaching handlers:

```go
// 401 Unauthorized — missing or invalid Bearer token
// Returned by pkg/echomiddleware/auth.go
{
  "message": "Unauthorized"
}

// 403 Forbidden — valid token but insufficient role
// Returned by pkg/echomiddleware/authorize.go
{
  "message": "Forbidden"
}

// 429 Too Many Requests — rate limit exceeded
// Includes Retry-After header with seconds until reset
{
  "message": "Too Many Requests"
}
```

Key rules:
- Return `401` when the token is absent, malformed, expired, or has an invalid signature
- Return `403` when the token is valid but the user lacks the required role or ownership
- Never expose token validation details in the error message (security)

### HTTP Layer Error Mapping
```go
// internal/application/users/handlers.go
func (h *Handler) CreateUser(c echo.Context) error {
    user, err := h.createHandler.CreateUser(ctx, req)
    if err != nil {
        return h.handleError(c, err)
    }

    return c.JSON(http.StatusCreated, user)
}

func (h *Handler) handleError(c echo.Context, err error) error {
    var createErr *CreateUserError
    if errors.As(err, &createErr) {
        switch createErr.Type {
        case "validation":
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Error: ErrorDetails{
                    Code:    "VALIDATION_ERROR",
                    Message: createErr.Message,
                    Details: extractValidationDetails(createErr.Cause),
                },
            })
        case "conflict":
            return c.JSON(http.StatusConflict, ErrorResponse{
                Error: ErrorDetails{
                    Code:    "USER_ALREADY_EXISTS",
                    Message: createErr.Message,
                },
            })
        case "internal":
            // Log full error for debugging
            slog.Error("Internal error creating user", "error", createErr.Cause)
            return c.JSON(http.StatusInternalServerError, ErrorResponse{
                Error: ErrorDetails{
                    Code:    "INTERNAL_ERROR",
                    Message: "Internal server error",
                },
            })
        }
    }

    // Fallback for unknown errors
    slog.Error("Unknown error", "error", err)
    return c.JSON(http.StatusInternalServerError, ErrorResponse{
        Error: ErrorDetails{
            Code:    "INTERNAL_ERROR",
            Message: "Internal server error",
        },
    })
}
```

## 🔍 Error Response Format

### Standard Error Structure
```go
type ErrorResponse struct {
    Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}
```

### Validation Error Details
```go
type ValidationDetails struct {
    Fields []FieldError `json:"fields"`
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   string `json:"value,omitempty"`
}
```

### Example Responses
```json
// Validation Error
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "fields": [
        {
          "field": "email",
          "message": "invalid email format",
          "value": "invalid-email"
        }
      ]
    }
  }
}

// Business Logic Error
{
  "error": {
    "code": "USER_ALREADY_EXISTS",
    "message": "User with this email already exists"
  }
}

// Internal Error (без деталей для безопасности)
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Internal server error"
  }
}
```

## 📊 Logging Strategy

### Structured Logging
```go
import "log/slog"

// Логируем с контекстом
slog.Error("Failed to create user",
    "error", err,
    "email", req.Email,
    "user_id", userID,
    "request_id", getRequestID(ctx),
)

// Логируем validation errors как warning
slog.Warn("Validation failed",
    "error", validationErr,
    "field", fieldName,
    "request_id", getRequestID(ctx),
)
```

### Log Levels
- **Error** - infrastructure failures, unexpected errors
- **Warn** - validation failures, business rule violations
- **Info** - successful operations, state changes
- **Debug** - detailed execution flow (только в development)

## 🔄 Error Recovery Patterns

### Retry Logic
```go
func (r *MongoUserRepository) CreateUser(ctx context.Context, email string, passwordHash []byte, createFn func() (*users.User, error)) (*users.User, error) {
    const maxRetries = 3

    for attempt := 0; attempt < maxRetries; attempt++ {
        user, err := r.attemptCreateUser(ctx, email, passwordHash, createFn)
        if err == nil {
            return user, nil
        }

        // Retry только для временных ошибок
        if !isRetryableError(err) {
            return nil, err
        }

        // Exponential backoff
        time.Sleep(time.Duration(attempt*100) * time.Millisecond)
    }

    return nil, fmt.Errorf("failed to create user after %d attempts", maxRetries)
}
```

### Circuit Breaker Pattern
```go
// Для внешних сервисов (email, notifications)
type CircuitBreaker struct {
    failures int
    lastFailure time.Time
    threshold int
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.isOpen() {
        return ErrCircuitBreakerOpen
    }

    err := fn()
    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.reset()
    return nil
}
```

## 🧪 Testing Error Scenarios

### Unit Tests for Error Cases
```go
func TestCreateUser_ValidationError(t *testing.T) {
    handler := NewCreateHandler(mockRepo)

    req := CreateUserRequest{
        Name:  "", // Invalid: empty name
        Email: "test@example.com",
    }

    _, err := handler.CreateUser(context.Background(), req)

    var createErr *CreateUserError
    require.True(t, errors.As(err, &createErr))
    assert.Equal(t, "validation", createErr.Type)
}
```

### Integration Tests with Real Errors
```go
func TestCreateUser_DatabaseError(t *testing.T) {
    // Симулируем недоступность БД
    container.Stop()

    _, err := handler.CreateUser(context.Background(), validRequest)

    var createErr *CreateUserError
    require.True(t, errors.As(err, &createErr))
    assert.Equal(t, "internal", createErr.Type)
}
```

---

> 💡 **Принцип**: Ошибки должны быть информативными для разработчиков, но безопасными для пользователей. Всегда логируем полный контекст ошибки для отладки.
