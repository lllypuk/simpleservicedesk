# Testing Strategy - Стратегия тестирования

## 🎯 Test Pyramid

### Test Levels
1. **Unit Tests** (70%) - изолированное тестирование бизнес-логики
2. **Integration Tests** (20%) - тестирование с реальными зависимостями
3. **End-to-End Tests** (10%) - полное тестирование пользовательских сценариев

### Test Distribution
```
     /\    E2E Tests
    /  \   API + Database + External Services
   /    \
  /______\  Integration Tests
 /        \ Repository + Database
/__________\ Unit Tests
Domain Logic + Application Logic
```

## 🧪 Unit Testing

### Domain Layer Testing
```go
// internal/domain/users/user_test.go
func TestNewUser_Success(t *testing.T) {
    // Test business logic without external dependencies
    user, err := users.NewUser("John Doe", "john@example.com")

    require.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)
    assert.Equal(t, users.RoleUser, user.Role) // Default role
    assert.NotEmpty(t, user.ID)
}

func TestNewUser_ValidationErrors(t *testing.T) {
    tests := []struct {
        name      string
        userName  string
        email     string
        wantError string
    }{
        {"empty name", "", "test@example.com", "name is required"},
        {"invalid email", "John", "invalid-email", "invalid email format"},
        {"empty email", "John", "", "email is required"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := users.NewUser(tt.userName, tt.email)

            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantError)
        })
    }
}
```

### Application Layer Testing with Mocks
```go
// internal/application/users_test/create_test.go
func TestCreateHandler_Success(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    handler := users.NewCreateHandler(mockRepo)

    req := users.CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "password123",
    }

    // Mock expectations
    mockRepo.On("CreateUser", mock.Anything, req.Email, mock.Anything, mock.Anything).
        Return(&users.User{
            ID:    uuid.New(),
            Name:  req.Name,
            Email: req.Email,
            Role:  users.RoleUser,
        }, nil)

    // Execute
    user, err := handler.CreateUser(context.Background(), req)

    // Verify
    require.NoError(t, err)
    assert.Equal(t, req.Name, user.Name)
    assert.Equal(t, req.Email, user.Email)
    mockRepo.AssertExpectations(t)
}
```

## 🔗 Integration Testing

### Database Integration Tests
```go
// test/integration/infrastructure/users/mongo_test.go
func TestMongoUserRepository_CreateUser(t *testing.T) {
    // Setup test container
    ctx := context.Background()
    mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mongo:7",
            ExposedPorts: []string{"27017/tcp"},
            WaitingFor:   wait.ForListeningPort("27017/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer mongoContainer.Terminate(ctx)

    // Get connection details
    host, err := mongoContainer.Host(ctx)
    require.NoError(t, err)
    port, err := mongoContainer.MappedPort(ctx, "27017")
    require.NoError(t, err)

    // Setup repository
    mongoURI := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
    repo, err := users.NewMongoRepository(mongoURI, "testdb")
    require.NoError(t, err)
    defer repo.Close()

    // Test data
    user := &users.User{
        Name:  "John Doe",
        Email: "john@example.com",
        Role:  users.RoleUser,
    }

    // Execute
    createdUser, err := repo.CreateUser(ctx, user.Email, []byte("hashed"), func() (*users.User, error) {
        return user, nil
    })

    // Verify
    require.NoError(t, err)
    assert.Equal(t, user.Name, createdUser.Name)
    assert.Equal(t, user.Email, createdUser.Email)
    assert.NotEmpty(t, createdUser.ID)

    // Verify in database
    foundUser, err := repo.GetUser(ctx, createdUser.ID)
    require.NoError(t, err)
    assert.Equal(t, createdUser.ID, foundUser.ID)
}
```

### API Integration Tests
```go
// test/integration/api/users_test.go
func TestUsersAPI_CreateAndGet(t *testing.T) {
    // Setup test server with real dependencies
    suite := setupTestSuite(t)
    defer suite.Cleanup()

    // Test data
    createReq := map[string]interface{}{
        "name":     "John Doe",
        "email":    "john@example.com",
        "password": "password123",
    }

    // Create user
    resp, err := suite.Client.Post("/users", createReq)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createResp users.CreateUserResponse
    err = json.NewDecoder(resp.Body).Decode(&createResp)
    require.NoError(t, err)

    // Get created user
    resp, err = suite.Client.Get(fmt.Sprintf("/users/%s", createResp.ID))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var user users.User
    err = json.NewDecoder(resp.Body).Decode(&user)
    require.NoError(t, err)
    assert.Equal(t, createReq["name"], user.Name)
    assert.Equal(t, createReq["email"], user.Email)
}
```

## 🔐 Auth Testing

Authentication and authorization are tested at multiple levels:

### Unit Tests
- `internal/application/auth/` — login handler, token generation, token validation

### Integration Tests (`test/integration/api/`)
- `POST /login` — valid credentials return JWT; invalid return 401
- Protected endpoints without token → 401
- Protected endpoints with wrong role → 403
- Rate limiting on `POST /login` — 6th request returns 429

### E2E Tests (`test/integration/e2e/`)
- User role change workflow: token issued with old role is rejected after role change
- Token refresh behavior: user must re-login after role change

## 🔄 E2E Tests

End-to-end tests verify complete user workflows:

```
test/integration/e2e/
├── ticket_lifecycle_test.go   # Open → InProgress → Resolved → Closed + timestamps
├── user_management_test.go    # Role changes, token refresh behavior
├── organization_test.go       # Org-scoped ticket filtering
├── category_test.go           # Tree changes, ticket association after parent move
└── error_scenarios_test.go    # Invalid transitions, duplicate emails, circular refs, 404s
```

Build tags: `//go:build integration,e2e`

Run with: `make test-e2e`

## 🏗️ Test Infrastructure

### Test Suite Setup
```go
// test/integration/suite.go
type TestSuite struct {
    Server    *echo.Echo
    Client    *TestClient
    MongoDB   testcontainers.Container
    UserRepo  users.Repository
    cleanup   []func()
}

func SetupTestSuite(t *testing.T) *TestSuite {
    ctx := context.Background()

    // Setup MongoDB container
    mongoContainer := setupMongoContainer(t, ctx)

    // Setup repositories
    userRepo := setupUserRepository(t, mongoContainer)

    // Setup application
    app := application.New(userRepo)
    server := setupServer(app)

    // Setup test client
    client := NewTestClient(server)

    suite := &TestSuite{
        Server:   server,
        Client:   client,
        MongoDB:  mongoContainer,
        UserRepo: userRepo,
    }

    // Register cleanup
    t.Cleanup(func() {
        suite.Cleanup()
    })

    return suite
}

func (s *TestSuite) Cleanup() {
    for _, cleanup := range s.cleanup {
        cleanup()
    }
}
```

### Test Data Factories
```go
// internal/testdata/factories.go
func NewUserFactory() *UserFactory {
    return &UserFactory{}
}

type UserFactory struct {
    name  string
    email string
    role  users.Role
}

func (f *UserFactory) WithName(name string) *UserFactory {
    f.name = name
    return f
}

func (f *UserFactory) WithEmail(email string) *UserFactory {
    f.email = email
    return f
}

func (f *UserFactory) WithRole(role users.Role) *UserFactory {
    f.role = role
    return f
}

func (f *UserFactory) Build() *users.User {
    return &users.User{
        ID:    uuid.New(),
        Name:  f.getOrDefault(f.name, "Test User"),
        Email: f.getOrDefault(f.email, "test@example.com"),
        Role:  f.getOrDefault(f.role, users.RoleUser),
    }
}

// Usage in tests
func TestCreateUser(t *testing.T) {
    user := NewUserFactory().
        WithName("Admin User").
        WithRole(users.RoleAdmin).
        Build()

    // Test with user...
}
```

## 📊 Test Coverage

### Coverage Goals
- **Domain Layer**: 90%+ - критическая бизнес-логика
- **Application Layer**: 85%+ - use cases и координация
- **Infrastructure Layer**: 70%+ - интеграции
- **Overall Project**: 80%+

### Coverage Commands
```bash
# Generate coverage report
make coverage_report

# View coverage by package
go test -cover ./internal/...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage Analysis
```go
//go:build coverage
// +build coverage

// Файлы, исключенные из coverage
// - Generated code (generated/*)
// - Test helpers (*_test.go)
// - Main functions (cmd/*)
```

## 🚀 Performance Testing

### Benchmark Tests
```go
// internal/domain/users/user_benchmark_test.go
func BenchmarkNewUser(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := users.NewUser("John Doe", "john@example.com")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkValidateEmail(b *testing.B) {
    emails := []string{
        "valid@example.com",
        "another@test.org",
        "user@domain.net",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        email := emails[i%len(emails)]
        validateEmail(email)
    }
}
```

### Load Testing (Future)
```go
// test/integration/load/users_test.go
func TestUsersAPI_Load(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }

    suite := setupTestSuite(t)
    defer suite.Cleanup()

    const (
        concurrency = 50
        requests    = 1000
    )

    // Implement load testing logic
    // Measure response times, throughput, error rates
}
```

## 🔧 Test Utilities

### Test Helpers
```go
// internal/testutil/helpers.go
func RequireValidUser(t *testing.T, user *users.User) {
    require.NotNil(t, user)
    require.NotEmpty(t, user.ID)
    require.NotEmpty(t, user.Name)
    require.NotEmpty(t, user.Email)
    require.NotZero(t, user.CreatedAt)
}

func AssertUsersEqual(t *testing.T, expected, actual *users.User) {
    assert.Equal(t, expected.ID, actual.ID)
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Email, actual.Email)
    assert.Equal(t, expected.Role, actual.Role)
}
```

### Test Configuration
```go
// testconfig/config.go
func GetTestConfig() *Config {
    return &Config{
        Database: DatabaseConfig{
            Driver: "memory",
        },
        Logger: LoggerConfig{
            Level: "debug",
            Output: "test", // Capture logs in tests
        },
    }
}
```

## 🎭 Mocking Strategy

### Repository Mocks
```go
// internal/application/mocks/user_repository.go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email string, passwordHash []byte, createFn func() (*users.User, error)) (*users.User, error) {
    args := m.Called(ctx, email, passwordHash, createFn)
    return args.Get(0).(*users.User), args.Error(1)
}

func (m *MockUserRepository) GetUser(ctx context.Context, id uuid.UUID) (*users.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*users.User), args.Error(1)
}

// Helper methods for common setup
func (m *MockUserRepository) ExpectCreateUser(user *users.User, err error) *mock.Call {
    return m.On("CreateUser", mock.Anything, user.Email, mock.Anything, mock.Anything).
        Return(user, err)
}
```

## 📝 Test Organization

### Test Commands
```makefile
# Makefile targets
test:
	go test -v ./internal/...

benchmark:
	go test -bench=. ./internal/...
```

### Test Naming Conventions
```
TestFunctionName_Scenario_ExpectedResult

Examples:
- TestCreateUser_ValidInput_ReturnsUser
- TestCreateUser_InvalidEmail_ReturnsValidationError
- TestCreateUser_DuplicateEmail_ReturnsConflictError
- TestGetUser_ExistingID_ReturnsUser
- TestGetUser_NonExistentID_ReturnsNotFoundError
```

---

> 💡 **Принцип**: Тесты должны быть быстрыми, независимыми, повторяемыми и детерминированными. Каждый тест проверяет одну конкретную функциональность.
