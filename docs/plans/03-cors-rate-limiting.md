# Plan: Add CORS and Rate Limiting

## Overview
Configure CORS to allow frontend applications to interact with the API.
Add rate limiting to protect against abuse.

## Validation Commands
- `make test`
- `make lint`

---

### Task 1: Add CORS middleware
- [ ] Configure Echo's built-in CORS middleware in `http_server.go`
- [ ] Set allowed origins (configurable via env `CORS_ALLOWED_ORIGINS`, default `*`)
- [ ] Set allowed methods: GET, POST, PUT, DELETE, OPTIONS
- [ ] Set allowed headers: Content-Type, Authorization
- [ ] Expose headers: X-Request-Id
- [ ] Add config fields to `internal/config.go`
- [ ] Mark completed

### Task 2: Add rate limiting
- [ ] Use Echo's built-in rate limiter middleware or `golang.org/x/time/rate`
- [ ] Configure global rate limit (configurable via env `RATE_LIMIT_RPS`, default 100)
- [ ] Apply stricter limit to `POST /login` (e.g., 5 req/min per IP) to prevent brute force
- [ ] Return `429 Too Many Requests` with `Retry-After` header
- [ ] Add config fields to `internal/config.go`
- [ ] Mark completed

### Task 3: Add tests
- [ ] Test CORS preflight requests return correct headers
- [ ] Test rate limiter returns 429 when exceeded
- [ ] Test login endpoint has stricter rate limit
- [ ] Mark completed

### Task 4: Update configuration docs
- [ ] Add new env variables to `.env.example`
- [ ] Update `README.md` environment variables table
- [ ] Mark completed
