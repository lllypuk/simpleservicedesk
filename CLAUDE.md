# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Development
- `make run` - Start the server locally on port 8080
- `make generate` - Generate server/client code from OpenAPI specs (required after API changes)
- `make lint` - Format code, organize imports, run golangci-lint, and verify generated code is up-to-date
- `make docker-up` - Start full stack with MongoDB using Docker Compose
- `make docker-down` - Stop and remove Docker containers
- `make docker-logs` - View logs from Docker containers
- `make docker-clean` - Clean up containers, volumes, and system cache
- `make docker-rebuild` - Full rebuild and restart with clean state

### Testing
- `make test` - Run unit tests only (excludes integration tests)
- `make test-unit` - Run unit tests explicitly
- `make test-integration` - Run all integration tests
- `make test-api` - Run HTTP API integration tests only
- `make test-repositories` - Run repository integration tests only
- `make test-e2e` - Run end-to-end tests only
- `make test-all` - Run both unit and integration tests
- `make coverage_report` - Generate HTML coverage report for unit tests
- `make coverage_integration` - Generate HTML coverage report for integration tests

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
- Repository implementations (MongoDB)
- External service integrations

### Generated Code (`generated/`)
Auto-generated from OpenAPI specs using oapi-codegen:
- Server interfaces and handlers
- Client code
- Type definitions
- OpenAPI specifications

## Key Technologies

- **Web Framework**: Echo v4 for HTTP routing and middleware
- **Storage**: MongoDB (primary)
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

## Current Development Status

The project is currently implementing REST API endpoints for all core entities:

### Completed Phases
- ✅ **Phase 1**: OpenAPI specification extended with all endpoints
- ✅ **Phase 2**: Tickets API - Full CRUD operations with status transitions
- ✅ **Phase 3**: Organizations API - Hierarchical organization management

### Current Phase
- 🚧 **Phase 4**: Categories API - Implementing category management with tree structure
  - Current focus: Category CRUD operations with parent-child relationships
  - See `.memory_bank/current_task.md` for detailed requirements

### Upcoming Phases
- 📋 **Phase 5**: Extended Users API - Additional user management operations

### API Status Summary
- **Users API**: ✅ Basic operations (create, get by ID) + 🚧 Extended operations pending
- **Tickets API**: ✅ Complete CRUD with status management
- **Organizations API**: ✅ Complete CRUD with hierarchical support  
- **Categories API**: 🚧 In progress - CRUD with tree structure
- **Generated Code**: ✅ Up-to-date with current OpenAPI spec

## Test Organization

The project uses a centralized integration test structure for better organization and maintenance:

### Test Directory Structure

```
test/
├── integration/
│   ├── api/           # HTTP API integration tests
│   ├── repositories/  # Database repository tests with testcontainers
│   ├── e2e/          # End-to-end workflow tests
│   └── shared/       # Common test utilities and setup
internal/
├── domain/           # Domain unit tests (co-located)
├── application/      # Application unit tests (co-located)
└── infrastructure/   # Infrastructure unit tests (co-located)
```

### Test Categories

**Unit Tests** (`internal/` packages):
- Domain logic validation
- Business rule verification
- Individual component testing
- Fast execution, no external dependencies
- Use mocks for dependencies

**Integration Tests** (`test/integration/`):
- API endpoint testing (`api/`)
- Database operations (`repositories/`)
- Component interaction verification
- Use testcontainers for real databases
- Tagged with `//go:build integration`

**End-to-End Tests** (`test/integration/e2e/`):
- Full workflow testing
- User journey simulation
- Multiple service interaction
- Tagged with `//go:build integration,e2e`

### Test Build Tags

All integration tests use build tags to separate them from unit tests:

```go
//go:build integration
// +build integration
```

This allows running different test types independently using Make commands.

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

**Test Package Naming Convention:**
- **CRITICAL**: All test files must use `package packagename_test` instead of `package packagename`
- This enforces black-box testing and prevents import cycles
- Tests should only use public APIs, not internal implementation details

**Example of correct test package naming:**
```go
// ❌ INCORRECT - Don't do this
package tickets
func TestInternalFunction(t *testing.T) { ... }

// ✅ CORRECT - Always use this
package tickets_test
import "myproject/internal/infrastructure/tickets"
func TestPublicAPI(t *testing.T) { 
    repo := tickets.NewMongoRepo(db) // Only test public API
}
```

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