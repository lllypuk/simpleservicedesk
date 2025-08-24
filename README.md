# Simple ServiceDesk

[![Go Reference](https://pkg.go.dev/badge/simpleservicedesk)](https://pkg.go.dev/simpleservicedesk) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A simple yet comprehensive ServiceDesk web service built with Go, providing RESTful APIs for managing users, tickets,
organizations, and categories in a service desk environment.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Configuration](#configuration)
    - [Code Generation](#code-generation)
    - [Running the Application](#running-the-application)
- [Usage Examples](#usage-examples)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Performance Profiling](#performance-profiling)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Overview

Simple ServiceDesk is a modern, minimalistic web service built with Go that provides essential functionality for
managing a service desk environment. The system supports user management, ticket tracking, organizational structures,
and categorization systems through a clean RESTful API.

## Features

### Core Functionality ✅ FULLY IMPLEMENTED
- **User Management**: Complete CRUD operations with role-based access control (Admin/Agent/User)
- **Ticket System**: Full ticket lifecycle management with status transitions and priority levels
- **Organization Management**: Hierarchical organizational structures with user relationships  
- **Category System**: Tree-structured categories for ticket classification
- **Comments System**: Rich commenting system for tickets with user attribution

### API & Architecture ✅ PRODUCTION READY
- **RESTful API**: Complete REST APIs following OpenAPI 3.0 specification
- **Code Generation**: Automatic server and client code generation from OpenAPI specs
- **Clean Architecture**: Domain-driven design with clear layer separation
- **MongoDB Support**: Full MongoDB integration with optimized queries
- **Comprehensive Testing**: Unit tests, integration tests, and API tests with testcontainers

### Operations & Infrastructure ✅ DEPLOYMENT READY
- **Structured Logging**: Comprehensive logging using Go's structured logging (slog)
- **Graceful Shutdown**: Proper signal handling and graceful application termination
- **Containerization**: Full Docker and Docker Compose support with optimized builds
- **Performance Profiling**: Built-in CPU and memory profiling capabilities
- **Development Tools**: Complete toolchain with linting, formatting, and code quality checks

## Architecture

The project follows clean architecture principles and is organized into the following layers:

- **Domain Layer** (`internal/domain/`): Contains core business entities and domain logic
    - Users with role management
    - Tickets with status and priority tracking
    - Organizations with hierarchical structures
    - Categories with nested relationships
- **Application Layer** (`internal/application/`): Coordinates business logic and use cases
- **Infrastructure Layer** (`internal/infrastructure/`): Implements external dependencies (databases, APIs)
- **Generated Code** (`generated/`): Auto-generated code from OpenAPI specifications
- **Presentation Layer**: HTTP handlers and middleware

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.24 or higher)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) (for code formatting)

### Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/your-username/simpleservicedesk.git
cd simpleservicedesk
go mod download
```

### Configuration

The application is configured using environment variables. Create a `.env` file in the project root:

```bash
# Application Environment
APP_ENV=development

# HTTP Server Configuration
HTTP_SERVER_PORT=8080

# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=servicedesk
```

### Code Generation

Generate server and client code from OpenAPI specifications:

```bash
make generate
```

This command will generate:

- Server interfaces and handlers
- Client code for API consumption
- Type definitions from OpenAPI schemas

### Running the Application

#### Local Development

Start the server locally:

```bash
make run
```

The server will be available at `http://localhost:8080`.

#### Using Docker Compose

Start the complete stack (application + MongoDB) using Docker:

```bash
docker-compose up -d
```

This will start:

- The SimpleServiceDesk application on port 8080
- MongoDB instance on port 27017
- Persistent volume for MongoDB data

## Usage Examples

### User Management

#### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john.doe@example.com",
    "password": "securepassword123"
  }'
```

#### Get User by ID

```bash
curl -X GET http://localhost:8080/users/{userId}
```

### Complete API Coverage

**All major API endpoints are fully implemented and tested:**

#### Users API ✅
- POST `/users` - Create user
- GET `/users/{id}` - Get user by ID  
- GET `/users` - List users
- PUT `/users/{id}` - Update user
- DELETE `/users/{id}` - Delete user
- PUT `/users/{id}/role` - Update user role
- GET `/users/{id}/tickets` - Get user's tickets

#### Tickets API ✅  
- POST `/tickets` - Create ticket
- GET `/tickets/{id}` - Get ticket by ID
- GET `/tickets` - List tickets  
- PUT `/tickets/{id}` - Update ticket
- DELETE `/tickets/{id}` - Delete ticket
- PUT `/tickets/{id}/status` - Update ticket status
- PUT `/tickets/{id}/assign` - Assign ticket to user
- POST `/tickets/{id}/comments` - Add comment
- GET `/tickets/{id}/comments` - Get comments

#### Organizations API ✅
- POST `/organizations` - Create organization  
- GET `/organizations/{id}` - Get organization by ID
- GET `/organizations` - List organizations
- PUT `/organizations/{id}` - Update organization
- DELETE `/organizations/{id}` - Delete organization
- GET `/organizations/{id}/users` - Get organization users
- GET `/organizations/{id}/tickets` - Get organization tickets

#### Categories API ✅
- POST `/categories` - Create category
- GET `/categories/{id}` - Get category by ID
- GET `/categories` - List categories  
- PUT `/categories/{id}` - Update category
- DELETE `/categories/{id}` - Delete category

## API Documentation

The API is documented using OpenAPI 3.0 specification. The specification file is located at `api/openapi.yaml`.

You can view the interactive API documentation by:

1. Running the application
2. Visiting the OpenAPI specification endpoint (if implemented)
3. Using tools like Swagger UI with the `api/openapi.yaml` file

## Testing ✅ COMPREHENSIVE COVERAGE

The project features comprehensive testing with multiple test types and strategies:

### Test Categories

#### Unit Tests
```bash
make test-unit
```
- Domain entity validation and business logic
- Application layer handlers and use cases
- Individual component testing with mocks
- Fast execution, no external dependencies

#### Integration Tests  
```bash
make test-integration
```
- HTTP API endpoint testing with real server
- MongoDB repository operations with testcontainers
- Component interaction verification
- Database schema validation

#### End-to-End Tests
```bash
make test-e2e  
```
- Complete user workflow simulation
- Multi-service interaction testing
- Real environment validation

### All Tests
```bash
make test-all      # Run both unit and integration tests
make test         # Run unit tests only (default)
```

### Coverage Reports
```bash
make coverage_report           # Unit test coverage 
make coverage_integration     # Integration test coverage
```

Both commands generate HTML coverage reports and automatically open them in your browser.

### Test Infrastructure
- **testcontainers-go**: Real MongoDB instances for integration testing
- **testify**: Assertions and test suites
- **Centralized test structure**: Organized in `test/integration/` for better maintenance

## Performance Profiling

The project includes built-in profiling capabilities:

### CPU Profiling

```bash
make cpu_profile
```

### Memory Profiling

```bash
make mem_profile
```

Both commands will start a web server at `http://localhost:6061` with interactive profiling data.

## Development

### Code Quality

The project maintains high code quality standards:

```bash
# Format code and run linters
make lint
```

This command will:

- Format code using `go fmt`
- Organize imports with `goimports`
- Run `golangci-lint` for comprehensive linting
- Verify that generated code is up to date

### Current Project Structure

```
├── api/                    # OpenAPI specifications
│   └── openapi.yaml       # Complete API specification
├── cmd/server/             # Application entry point
├── generated/              # Auto-generated code from OpenAPI
│   └── openapi/           # Server interfaces, types, client
├── internal/               # Private application code
│   ├── application/        # ✅ HTTP handlers, use cases
│   │   ├── users/         # User management handlers
│   │   ├── tickets/       # Ticket management handlers  
│   │   ├── organizations/ # Organization handlers
│   │   └── categories/    # Category handlers
│   ├── domain/            # ✅ Business entities and logic
│   │   ├── users/         # User domain model
│   │   ├── tickets/       # Ticket domain model
│   │   ├── organizations/ # Organization domain model
│   │   └── categories/    # Category domain model
│   └── infrastructure/    # ✅ External dependencies
│       ├── users/         # MongoDB user repository
│       ├── tickets/       # MongoDB ticket repository
│       ├── organizations/ # MongoDB organization repository
│       └── categories/    # MongoDB category repository
├── test/integration/       # ✅ Centralized integration tests
│   ├── api/               # HTTP API tests
│   ├── repositories/      # Database tests
│   ├── e2e/              # End-to-end tests
│   └── shared/           # Common test utilities
├── pkg/                   # Public packages (middleware, utilities)
└── profiles/              # Performance profiling data
```

### Environment Variables

| Variable           | Description               | Default                     |
|--------------------|---------------------------|-----------------------------|
| `APP_ENV`          | Application environment   | `development`               |
| `HTTP_SERVER_PORT` | HTTP server port          | `8080`                      |
| `MONGO_URI`        | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DATABASE`   | MongoDB database name     | `servicedesk`               |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to
discuss what you would like to change.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test && make lint`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
