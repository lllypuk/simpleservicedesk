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
- [ ] Create `internal/application/health/` package
- [ ] Implement `CheckMongoDB(ctx)` — ping MongoDB, return status + latency
- [ ] Define response types: status (healthy/unhealthy), checks array, version info
- [ ] Mark completed

### Task 2: Add health check endpoints
- [ ] `GET /health/live` — simple liveness (always 200 if process is running)
- [ ] `GET /health/ready` — readiness (checks MongoDB connectivity)
- [ ] Return structured JSON: `{"status": "healthy", "checks": [{"name": "mongodb", "status": "up", "latency_ms": 5}]}`
- [ ] Return 503 if any check fails
- [ ] Keep `/ping` for backwards compatibility
- [ ] Mark completed

### Task 3: Add to Docker Compose
- [ ] Add `healthcheck` section to `docker-compose.yml` for the app service
- [ ] Use `GET /health/ready` as health check command
- [ ] Mark completed

### Task 4: Add tests
- [ ] Unit test health handler with mock DB ping
- [ ] Integration test with real MongoDB
- [ ] Mark completed
