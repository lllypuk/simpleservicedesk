# Plan: Add JWT Authentication & Authorization

## Overview
Add simple JWT-based authentication directly in the service.
Admin creates users (with passwords already in domain). Users authenticate via
`POST /login`, receive a JWT token, and use it for all subsequent requests.
Authorization middleware checks user roles against endpoint requirements.

## Validation Commands
- `make test`
- `make lint`
- `make test-integration`

## Design Decisions
- **JWT library**: `golang-jwt/jwt/v5` (industry standard for Go)
- **Token storage**: Stateless — JWT contains user ID and role, validated by signature
- **Password hashing**: Already implemented with bcrypt in domain layer
- **Token lifetime**: Configurable, default 24h
- **Refresh tokens**: Out of scope for v1, can be added later
- **Registration**: No self-registration — only Admin can create users (existing `POST /users`)

## Route Protection Matrix
| Endpoint | Public | User | Agent | Admin |
|----------|--------|------|-------|-------|
| `POST /login` | x | | | |
| `GET /users` | | | x | x |
| `POST /users` | | | | x |
| `PUT /users/{id}/role` | | | | x |
| `DELETE /users/{id}` | | | | x |
| `GET /tickets` | | x* | x | x |
| `POST /tickets` | | x | x | x |
| `PUT /tickets/{id}/status` | | | x | x |
| `POST /tickets/{id}/assign` | | | x | x |
| Other endpoints | | x | x | x |

*User sees only own tickets

---

### Task 1: Add JWT dependency
- [ ] Add `github.com/golang-jwt/jwt/v5` to `go.mod`
- [ ] Run `go mod tidy`
- [ ] Mark completed

### Task 2: Add auth configuration
- [ ] Add `JWT_SECRET` and `JWT_EXPIRATION` to config (`internal/config.go`)
- [ ] Add to `.env.example`
- [ ] Generate a default secret for development
- [ ] Mark completed

### Task 3: Create auth domain types
- [ ] Create `internal/domain/auth/` package
- [ ] Define `Claims` struct (UserID, Role, standard JWT claims)
- [ ] Define `LoginRequest` / `LoginResponse` types
- [ ] Add unit tests
- [ ] Mark completed

### Task 4: Create auth service
- [ ] Create `internal/application/auth/` package
- [ ] Implement `Login(email, password)` — validate credentials, return JWT
- [ ] Implement `GenerateToken(user)` — create signed JWT
- [ ] Implement `ValidateToken(tokenString)` — parse and validate JWT
- [ ] Add unit tests with mocked user repository
- [ ] Mark completed

### Task 5: Create login endpoint
- [ ] Add `POST /login` to `api/openapi.yaml` (LoginRequest, LoginResponse schemas)
- [ ] Run `make generate`
- [ ] Implement login handler in `internal/application/auth/login.go`
- [ ] Register route in `http_server.go`
- [ ] Add unit tests
- [ ] Mark completed

### Task 6: Create auth middleware
- [ ] Create `pkg/echomiddleware/auth.go`
- [ ] Implement JWT extraction from `Authorization: Bearer <token>` header
- [ ] Parse token, inject user claims into echo.Context
- [ ] Return 401 for missing/invalid tokens
- [ ] Add unit tests
- [ ] Mark completed

### Task 7: Create authorization middleware
- [ ] Create `pkg/echomiddleware/authorize.go`
- [ ] Implement role-checking middleware: `RequireRole(minRole Role)`
- [ ] Implement ownership check helper: `IsOwnerOrRole(userID, minRole)`
- [ ] Add unit tests
- [ ] Mark completed

### Task 8: Apply middleware to routes
- [ ] Update `http_server.go` — group routes by access level
- [ ] Public routes: `POST /login`, `GET /ping`
- [ ] Authenticated routes: wrap with auth middleware
- [ ] Admin-only routes: wrap with `RequireRole(Admin)`
- [ ] Agent+ routes: wrap with `RequireRole(Agent)`
- [ ] Ensure `POST /users` requires Admin role
- [ ] Mark completed

### Task 9: Update existing handlers for auth context
- [ ] Update ticket handlers — filter by user's own tickets for User role
- [ ] Update user handlers — restrict `DELETE`, `PUT /role` to Admin
- [ ] Extract current user from context in handlers that need it
- [ ] Mark completed

### Task 10: Update OpenAPI specification
- [ ] Add `securitySchemes` (bearerAuth) to `api/openapi.yaml`
- [ ] Add `security` requirements to protected endpoints
- [ ] Run `make generate`
- [ ] Mark completed

### Task 11: Integration tests
- [ ] Add auth helper to `test/integration/shared/` — login and get token
- [ ] Add tests for `POST /login` (success, wrong password, nonexistent user)
- [ ] Add tests for protected endpoint without token (expect 401)
- [ ] Add tests for insufficient role (expect 403)
- [ ] Add tests for User role seeing only own tickets
- [ ] Update existing integration tests to include auth headers
- [ ] Mark completed

### Task 12: Documentation
- [ ] Update `README.md` — add authentication section with usage examples
- [ ] Update `CLAUDE.md` — mention auth architecture
- [ ] Update `docs/tech_stack.md` — add JWT dependency
- [ ] Mark completed
