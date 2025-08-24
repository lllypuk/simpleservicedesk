# Integration Tests with TestContainers

This directory contains integration tests that use real MongoDB databases via TestContainers instead of mocks.

## Overview

The integration tests are organized into the following structure:

```
test/integration/
├── api/                 # HTTP API endpoint tests
├── repositories/        # Database repository tests
├── e2e/                # End-to-end workflow tests (future)
└── shared/             # Common test utilities and setup
```

## TestContainers Setup

All integration tests now use **TestContainers** with real MongoDB instances instead of mock repositories. This provides:

- **Real database behavior**: Tests run against actual MongoDB operations
- **Isolation**: Each test suite gets its own MongoDB container
- **Consistency**: Identical behavior between tests and production
- **Reliability**: Tests catch database-specific issues that mocks might miss

### MongoDB Container

- **Image**: `mongodb/mongodb-community-server:8.0-ubi8`
- **Database**: `test_integration` (created per test suite)
- **Lifecycle**: Containers are automatically created, started, and cleaned up

## Test Categories

### API Tests (`api/`)

Test HTTP endpoints with real database backends:

```bash
make test-api
```

**Features tested:**
- Complete request/response cycles
- Database persistence verification
- Error handling with real database errors
- Data validation with actual storage constraints

### Repository Tests (`repositories/`)

Test database repository implementations directly:

```bash
make test-repositories
```

**Features tested:**
- CRUD operations
- Complex queries and filtering
- Transaction behavior
- Concurrent access patterns
- Error handling and edge cases

## Running Tests

### All Integration Tests
```bash
make test-integration
```

### Specific Test Categories
```bash
make test-api           # API endpoint tests only
make test-repositories  # Repository tests only
make test-e2e          # End-to-end tests only (when available)
```

### Test Requirements

- **Docker**: Required for TestContainers
- **Build Tags**: Tests use `//go:build integration` tag
- **Isolation**: Each test clears database collections before running

## Test Structure

### Shared Setup (`shared/setup.go`)

The `IntegrationSuite` provides common setup for all integration tests:

```go
type IntegrationSuite struct {
    suite.Suite
    HTTPServer        *echo.Echo
    UsersRepo         application.UserRepository
    TicketsRepo       application.TicketRepository
    OrganizationsRepo application.OrganizationRepository
    MongoContainer    *mongodb.MongoDBContainer
    MongoDB           *mongo.Database
    MongoClient       *mongo.Client
}
```

**Key Features:**
- Automatic MongoDB container lifecycle management
- Real repository instances (no mocks)
- Clean database state between tests
- HTTP server setup with real repositories

### Test Lifecycle

1. **Suite Setup**: MongoDB container starts, repositories initialize
2. **Test Setup**: Database collections are cleared
3. **Test Execution**: Test runs with clean database state
4. **Suite Teardown**: Container stops and is removed

## Migration from Mocks

The integration tests were migrated from mock repositories to TestContainers:

### Before (Mocks)
```go
// Mock repositories with in-memory state
s.UsersRepo = newMockUserRepository()
s.TicketsRepo = newMockTicketRepository()
```

### After (TestContainers)
```go
// Real MongoDB repositories with TestContainers
s.UsersRepo = users.NewMongoRepo(s.MongoDB)
s.TicketsRepo = tickets.NewMongoRepo(s.MongoDB)
```

### Benefits of Migration

1. **Real Database Behavior**: Tests catch MongoDB-specific issues
2. **Data Persistence**: Verify actual database operations
3. **Query Testing**: Test complex filters and aggregations
4. **Performance Testing**: Measure real database performance
5. **Error Handling**: Test actual database error conditions

## Best Practices

### Writing New Tests

1. **Use IntegrationSuite**: Extend `shared.IntegrationSuite` for automatic setup
2. **Clean State**: Tests automatically get clean database state
3. **Real Data**: Create realistic test data that exercises business logic
4. **Error Cases**: Test actual database error conditions

### Test Performance

- **Parallel Execution**: Tests within a suite run sequentially, but different suites can run in parallel
- **Container Reuse**: Each test suite reuses the same container across test methods
- **Fast Cleanup**: Database collections are dropped, not recreated

### Debugging

- **Container Logs**: Use `make docker-logs` if needed
- **Database Inspection**: Connect to test database during debugging
- **Verbose Output**: Use `-v` flag for detailed test output

## Dependencies

The following dependencies are required for TestContainers integration:

```go
github.com/testcontainers/testcontainers-go v0.38.0
github.com/testcontainers/testcontainers-go/modules/mongodb v0.38.0
go.mongodb.org/mongo-driver v1.17.4
```

## Troubleshooting

### Common Issues

1. **Docker Not Running**: Ensure Docker daemon is running
2. **Permission Issues**: Check Docker socket permissions
3. **Port Conflicts**: TestContainers automatically handles port allocation
4. **Container Cleanup**: Containers are automatically cleaned up, but check with `docker ps` if needed

### Performance

- **Cold Start**: First test run may be slower due to image download
- **Subsequent Runs**: Faster execution with cached images
- **Cleanup Time**: Allow time for container termination

## Future Enhancements

### Planned Improvements

1. **E2E Tests**: Complete workflow tests across multiple services
2. **Performance Benchmarks**: Standardized performance test suite
3. **Data Migration Tests**: Test database schema migrations
4. **Backup/Restore Tests**: Test data persistence scenarios

### Configuration Options

Future versions may support:
- Custom MongoDB versions
- Additional database configurations
- Shared containers across test suites
- Custom test data fixtures