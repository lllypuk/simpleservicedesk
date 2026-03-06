# Plan: Improve Health Check Endpoint

## Overview
Replace the simple `/ping` endpoint with proper health check endpoints
that verify MongoDB connectivity. Add readiness/liveness probes for
container orchestration (Docker, Kubernetes).

## Validation Commands
- `make test`
- `make lint`

---

### Task 1: Create health check service
- [x] Create `internal/application/health/` package
- [x] Implement `CheckMongoDB(ctx)` — ping MongoDB, return status + latency
- [x] Define response types: status (healthy/unhealthy), checks array, version info
- [x] Mark completed

### Task 2: Add health check endpoints
- [x] `GET /health/live` — simple liveness (always 200 if process is running)
- [x] `GET /health/ready` — readiness (checks MongoDB connectivity)
- [x] Return structured JSON: `{"status": "healthy", "checks": [{"name": "mongodb", "status": "up", "latency_ms": 5}]}`
- [x] Return 503 if any check fails
- [x] Keep `/ping` for backwards compatibility
- [x] Mark completed

### Task 3: Add to Docker Compose
- [x] Add `healthcheck` section to `docker-compose.yml` for the app service
- [x] Use `GET /health/ready` as health check command
- [x] Mark completed

### Task 4: Add tests
- [x] Unit test health handler with mock DB ping
- [x] Integration test with real MongoDB
- [x] Mark completed
