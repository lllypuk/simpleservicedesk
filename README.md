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

- **User Management**: Create, retrieve, and manage users with role-based access control
- **Ticket System**: Full ticket lifecycle management with status tracking and priority levels
- **Organization Management**: Handle organizational structures and hierarchies
- **Category System**: Organize tickets and resources with hierarchical categories
- **RESTful API**: Clean, well-documented APIs following OpenAPI 3.0 specification
- **Code Generation**: Automatic server and client code generation from OpenAPI specs
- **Multiple Storage Backends**: Support for both in-memory and MongoDB repositories
- **Structured Logging**: Comprehensive logging using Go's structured logging (slog)
- **Graceful Shutdown**: Proper signal handling and graceful application termination
- **Containerization**: Full Docker and Docker Compose support
- **Clean Architecture**: Follows clean architecture principles with clear separation of concerns

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

### Additional API Endpoints

The system provides additional endpoints for:

- Ticket management
- Organization operations
- Category management
- User role assignments

Refer to the [API Documentation](#api-documentation) for complete endpoint details.

## API Documentation

The API is documented using OpenAPI 3.0 specification. The specification file is located at `api/openapi.yaml`.

You can view the interactive API documentation by:

1. Running the application
2. Visiting the OpenAPI specification endpoint (if implemented)
3. Using tools like Swagger UI with the `api/openapi.yaml` file

## Testing

### Run All Tests

```bash
make test
```

### Unit Tests Only

```bash
make unit_test
```

### Integration Tests Only

```bash
make integration_test
```

### Coverage Report

```bash
make coverage_report
```

This will generate an HTML coverage report and automatically open it in your browser.

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

### Project Structure

```
├── api/                    # OpenAPI specifications
├── build/                  # Build configurations (Dockerfiles)
├── cmd/server/             # Application entry point
├── generated/              # Auto-generated code
├── integration_test/       # Integration tests
├── internal/               # Private application code
│   ├── application/        # Application layer
│   ├── domain/            # Domain layer
│   └── infrastructure/    # Infrastructure layer
├── pkg/                    # Public packages
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
