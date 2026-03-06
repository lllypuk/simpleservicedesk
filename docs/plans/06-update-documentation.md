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
- [x] Add new environment variables (JWT_SECRET, CORS, rate limiting)
- [x] Update development status section
- [x] Remove outdated "all completed" markers, replace with actual current status
- [x] Add auth-related make commands if any (N/A — no new commands needed)
- [x] Mark completed

### Task 2: Update README.md
- [x] Update prerequisites — Go 1.26
- [x] Add "Authentication" section with login flow and curl examples
- [x] Add Authorization header examples to all curl commands
- [x] Update environment variables table with new vars
- [x] Add health check endpoints to API documentation section
- [x] Update project structure if new packages were added
- [x] Remove excessive "FULLY IMPLEMENTED" / "PRODUCTION READY" markers
- [x] Mark completed

### Task 3: Update docs/tech_stack.md
- [x] Update Go version to 1.26
- [x] Add JWT library to dependencies
- [x] Add authentication section
- [x] Fix test directory path (`test/integration/` not `integration_test/`)
- [x] Update environment variables table
- [x] Mark completed

### Task 4: Update docs/product_brief.md
- [x] Update MVP section — authentication is now implemented
- [x] Move "Автоматическое назначение тикетов" from planned to current if implemented (N/A — not implemented)
- [x] Mark completed

### Task 5: Update docs/current_task.md
- [x] Replace categories test task with current development focus
- [x] Or remove if no active task
- [x] Mark completed

### Task 6: Update docs/testing_assessment_and_plan.md
- [x] Moved/deleted: old file is no longer present after docs migration
- [x] Add categories API tests to the active docs set (covered in guides/testing_strategy.md)
- [x] Add E2E tests section (added to guides/testing_strategy.md)
- [x] Add auth-related tests section (added to guides/testing_strategy.md)
- [x] Mark completed

### Task 7: Clean up docs/patterns/
- [x] Update `error_handling.md` — add auth error patterns (401, 403)
- [x] Update `api_standards.md` — add security/auth section
- [x] Mark completed

### Task 8: Create docs/plans/README.md
- [x] Create index file listing all plans with status
- [x] Include execution order and dependencies between plans
- [x] Mark completed
