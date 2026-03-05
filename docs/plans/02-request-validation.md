# Plan: Add OpenAPI Request Validation Middleware

## Overview
Add automatic request validation against the OpenAPI schema using
`oapi-codegen`'s built-in middleware for Echo. This ensures all incoming
requests are validated (required fields, types, formats, enums) before
reaching handlers, reducing boilerplate validation in handlers.

## Validation Commands
- `make test`
- `make lint`
- `make test-integration`

---

### Task 1: Add validation middleware dependency
- [ ] Add `github.com/oapi-codegen/echo-middleware` to `go.mod`
- [ ] Run `go mod tidy`
- [ ] Mark completed

### Task 2: Integrate validation middleware
- [ ] Load OpenAPI spec using `generated/openapi/spec.go` (GetSwagger)
- [ ] Create validation middleware in `http_server.go`
- [ ] Configure to skip validation for non-API routes (`/ping`, `/login`)
- [ ] Return structured error responses (consistent with existing error format)
- [ ] Mark completed

### Task 3: Clean up handler validation
- [ ] Review handlers for redundant validation now covered by middleware
- [ ] Remove duplicate checks (required fields, type validation)
- [ ] Keep business logic validation (domain rules not expressible in OpenAPI)
- [ ] Mark completed

### Task 4: Add tests
- [ ] Add unit test for validation middleware error responses
- [ ] Add integration tests: missing required fields, wrong types, invalid enums
- [ ] Verify existing tests still pass
- [ ] Mark completed

### Task 5: Fix the TODO
- [ ] Fix `internal/application/organizations/list.go:41` — add proper pagination response
- [ ] Add test for pagination response
- [ ] Mark completed
