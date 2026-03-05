# Plan: Update Documentation

## Overview
Bring all documentation up to date after implementing the changes above.
Fix outdated references, add new sections, and ensure consistency across
`README.md`, `CLAUDE.md`, and `docs/`.

## Validation Commands
- Review all changed files manually

## Prerequisites
- Should be done after all other plans are implemented (or iteratively as each completes)

---

### Task 1: Update CLAUDE.md
- [x] Update Go version reference to 1.26
- [x] Add authentication architecture section
- [ ] Add new environment variables (JWT_SECRET, CORS, rate limiting)
- [x] Update development status section
- [x] Remove outdated "all completed" markers, replace with actual current status
- [ ] Add auth-related make commands if any
- [ ] Mark completed

### Task 2: Update README.md
- [x] Update prerequisites — Go 1.26
- [x] Add "Authentication" section with login flow and curl examples
- [x] Add Authorization header examples to all curl commands
- [x] Update environment variables table with new vars
- [ ] Add health check endpoints to API documentation section
- [x] Update project structure if new packages were added
- [ ] Remove excessive "FULLY IMPLEMENTED" / "PRODUCTION READY" markers
- [ ] Mark completed

### Task 3: Update docs/tech_stack.md
- [ ] Update Go version to 1.26
- [ ] Add JWT library to dependencies
- [ ] Add authentication section
- [ ] Fix test directory path (`test/integration/` not `integration_test/`)
- [ ] Update environment variables table
- [ ] Mark completed

### Task 4: Update docs/product_brief.md
- [ ] Update MVP section — authentication is now implemented
- [ ] Move "Автоматическое назначение тикетов" from planned to current if implemented
- [ ] Mark completed

### Task 5: Update docs/current_task.md
- [ ] Replace categories test task with current development focus
- [ ] Or remove if no active task
- [ ] Mark completed

### Task 6: Update docs/testing_assessment_and_plan.md
- [ ] Moved/deleted: old file is no longer present after docs migration
- [ ] Add categories API tests to the active docs set
- [ ] Add E2E tests section
- [ ] Add auth-related tests section
- [ ] Mark completed

### Task 7: Clean up docs/patterns/
- [ ] Update `error_handling.md` — add auth error patterns (401, 403)
- [ ] Update `api_standards.md` — add security/auth section
- [ ] Mark completed

### Task 8: Create docs/plans/README.md
- [x] Create index file listing all plans with status
- [x] Include execution order and dependencies between plans
- [ ] Mark completed
