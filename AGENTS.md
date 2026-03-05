# AGENTS.md

Operational guide for autonomous coding agents working in this repository.

## Project Snapshot

- Language: Go (`go 1.26`)
- Module: `simpleservicedesk`
- Architecture: clean architecture (`internal/domain`, `internal/application`, `internal/infrastructure`)
- API contract: OpenAPI in `api/openapi.yaml`, generated code in `generated/openapi`
- Primary DB: MongoDB
- Integration tests: Testcontainers + Docker

## Authoritative Command Sources

- `Makefile` (primary source of runnable workflows)
- `CLAUDE.md` (repo-specific development rules)
- `.golangci.yml` (lint/style enforcement)
- `test/integration/README.md` (integration test behavior)
- `check-go-generate.sh` (generated code freshness check)

## Build, Lint, Test Commands

### Daily Development

- Run service locally: `make run`
- Generate code from OpenAPI: `make generate`
- Full lint/format/check pipeline: `make lint`
- Unit tests only (default): `make test` or `make test-unit`
- Unit + integration: `make test-all`

### Integration Test Targets

- All integration tests: `make test-integration`
- API integration tests only: `make test-api`
- Repository integration tests only: `make test-repositories`
- E2E tests only: `make test-e2e`

### Coverage and Profiling

- Unit coverage report: `make coverage_report`
- Integration coverage report: `make coverage_integration`
- CPU profiling: `make cpu_profile`
- Memory profiling: `make mem_profile`

### Docker Helpers

- Start stack: `make docker-up`
- Stop stack: `make docker-down`
- Tail logs: `make docker-logs`
- Clean containers/volumes/cache: `make docker-clean`
- Rebuild clean stack: `make docker-rebuild`

## Run a Single Test (Important)

Use `go test -run` with a package path. Prefer anchored regex for exact match.

- Exact single test in one package:
  - `go test -v -run '^TestUserValidation$' ./internal/domain/users`
- Match multiple related tests in one package:
  - `go test -v -run 'TestUser.*' ./internal/domain/users`
- Run one integration suite test:
  - `go test -v -tags=integration -run '^TestUserAPI$' ./test/integration/api`
- Run one method-style suite test (name match):
  - `go test -v -tags=integration -run 'TestCreateUserIntegration' ./test/integration/api`

Notes:

- `go test` runs at package scope, then filters by `-run` regex.
- Integration tests require `-tags=integration` (or `integration,e2e` for e2e).
- Docker must be available for Testcontainers-based integration tests.

## Lint/Format Details Enforced by Repo

`make lint` executes:

1. `go fmt ./...`
2. `goimports` on all Go files except `generated/*`
3. `golangci-lint run ./...`
4. `./check-go-generate.sh`

Implications:

- Keep imports organized by `goimports`.
- Do not manually edit generated files as source-of-truth changes.
- If lint fails due to generated diffs, run `make generate` and commit regenerated output.

## Code Style and Conventions

### Formatting and Imports

- Follow `gofmt` and `goimports` output exactly.
- Keep local imports under module prefix (`simpleservicedesk/...`).
- Prefer three import groups: stdlib, third-party, local module.
- Keep lines manageable (formatter/linter target is 120 chars).

### Naming

- Use clear English names for identifiers, comments, and test descriptions.
- Exported names: PascalCase; internal names: camelCase.
- Domain errors use `Err...` naming (for `errors.Is` checks).
- Keep names descriptive; avoid cryptic abbreviations.

### Types and APIs

- Prefer explicit types over ambiguous shorthand in public surfaces.
- Keep `context.Context` as first parameter for I/O or request-scoped operations.
- Preserve existing repository interface patterns in `internal/application/interfaces.go`.
- Respect domain invariants by constructing entities via domain constructors.

### Error Handling

- Never ignore returned errors.
- Wrap contextual errors with `%w` when propagating external/internal failures.
- Use `errors.Is` for sentinel comparison.
- Return domain-appropriate errors for validation/not-found/conflict paths.
- Avoid panic-driven control flow in application/infrastructure layers.

### Logging

- Use `log/slog` patterns used in repo.
- Prefer context-aware logging methods.
- Avoid global logger usage patterns (enforced by lint configuration).

### Testing Conventions

- Test packages should use external form: `package <name>_test`.
- Prefer table-driven tests for validation/branch coverage.
- Keep integration tests under `test/integration/...` with build tags.
- Use realistic fixtures and assert on behavior, not internals.
- For integration tests, rely on shared setup utilities in `test/integration/shared`.

### Generated Code and OpenAPI Workflow

- If `api/openapi.yaml` changes, always run `make generate`.
- Implement corresponding logic after generation.
- Finish with `make lint` and relevant tests.

## Lint-Driven Guardrails to Respect

From `.golangci.yml` (non-exhaustive, high impact):

- No unchecked errors (`errcheck`).
- No `init` functions (`gochecknoinits`).
- Avoid global mutable state (`gochecknoglobals`, `reassign`).
- Separate test package usage is enforced (`testpackage`).
- Enforce proper `nolint` usage with specific linter + explanation.
- Prefer `log/slog` over `log` in non-`main.go` files.
- Avoid deprecated package choices (see `depguard` rules).

## Agent Execution Checklist

When making changes, run this sequence unless task scope clearly dictates otherwise:

1. `make generate` (only if API/schema changed)
2. `make lint`
3. Targeted tests first (single package or `-run` pattern)
4. Broader suite (`make test` or `make test-all`) as appropriate

Before finalizing work:

- Ensure modified code and tests compile.
- Ensure lint passes cleanly.
- Ensure generated artifacts are up to date.
- Document any known pre-existing failures explicitly.

## Cursor/Copilot Rules Check

At time of writing, no repository-specific Cursor/Copilot instruction files were found:

- `.cursor/rules/**` not found
- `.cursorrules` not found
- `.github/copilot-instructions.md` not found

If these files are added later, treat them as higher-priority agent instructions and merge their guidance into this document.
