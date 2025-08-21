# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Development
- `make run` - Start the server locally on port 8080
- `make generate` - Generate server/client code from OpenAPI specs (required after API changes)
- `make lint` - Format code, organize imports, run golangci-lint, and verify generated code is up-to-date
- `docker-compose up -d` - Start full stack with MongoDB

### Testing
- `make test` - Run all tests (unit + integration)
- `make coverage_report` - Generate HTML coverage report

### Performance
- `make cpu_profile` - CPU profiling with web interface on :6061
- `make mem_profile` - Memory profiling with web interface on :6061

## Architecture

This is a Go service desk application following clean architecture principles:

### Domain Layer (`internal/domain/`)
Core business entities with their own validation and business logic:
- **Users**: Role-based access control (Admin, Agent, User)
- **Tickets**: Status transitions (Open → InProgress → Resolved → Closed) and priority levels
- **Organizations**: Hierarchical structures
- **Categories**: Nested categorization system

### Application Layer (`internal/application/`)
- Coordinates business logic and use cases
- Defines repository interfaces (e.g., `UserRepository` in `interfaces.go`)
- HTTP handlers and middleware

### Infrastructure Layer (`internal/infrastructure/`)
- Repository implementations (in-memory and MongoDB)
- External service integrations

### Generated Code (`generated/`)
Auto-generated from OpenAPI specs using oapi-codegen:
- Server interfaces and handlers
- Client code
- Type definitions
- OpenAPI specifications

## Key Technologies

- **Web Framework**: Echo v4 for HTTP routing and middleware
- **Storage**: MongoDB (primary) with in-memory fallback for testing
- **Code Generation**: oapi-codegen from OpenAPI 3.0 specs
- **Testing**: testcontainers-go for integration tests with real MongoDB
- **Logging**: Structured logging using Go's slog package

## Development Workflow

1. Make API changes in `api/openapi.yaml`
2. Run `make generate` to update generated code
3. Implement business logic in domain layer
4. Update application/infrastructure layers as needed
5. Run `make lint` before committing
6. Ensure all tests pass with `make test`

## Code Quality Requirements

**IMPORTANT**: After completing each task, you must run:

1. **`make lint`** - Code formatting, import organization, and linter checks
   - Automatically fixes formatting using `go fmt`
   - Organizes imports using `goimports`
   - Runs `golangci-lint` for code quality checks
   - Verifies generated code is up-to-date

2. **`make test`** - Run all tests to ensure functionality
   - Unit tests (`./internal/...`)
   - Integration tests (`./integration_test/...`) 
   - All tests must pass before committing

**Post-development workflow:**
```bash
make lint    # Fix all linter issues
make test    # Ensure all tests pass
git add .    # Only after successful lint + test
git commit
```

Code should NOT be committed if:
- `make lint` shows errors
- `make test` shows failing tests
- Generated code is not up-to-date

## Code Style Guidelines

**Language Requirements:**
- All code comments must be written in **English only**
- Variable names, function names, and identifiers should use English
- Documentation and README files can be in multiple languages
- Test descriptions and error messages should be in English

**Example of correct commenting:**
```go
// CreateUser creates a new user with the provided email and password
func CreateUser(email string, password string) (*User, error) {
    // Validate email format before processing
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    // ... rest of implementation
}
```

## Configuration

Uses environment variables (see `.env` file):
- `APP_ENV`: Application environment (development/production)
- `HTTP_SERVER_PORT`: Server port (default: 8080)
- `MONGO_URI`: MongoDB connection string
- `MONGO_DATABASE`: MongoDB database name

## Code Generation Dependencies

The project heavily relies on code generation. Always run `make generate` after:
- Modifying `api/openapi.yaml`
- Adding new API endpoints
- Changing request/response schemas

The `check-go-generate.sh` script (run by `make lint`) verifies generated code is current.