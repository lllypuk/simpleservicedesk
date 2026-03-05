# Plan: Add End-to-End Tests

## Overview
Implement E2E tests that verify complete user workflows — from login
to creating tickets, assigning agents, and resolving tickets. These tests
use the full stack (HTTP server + MongoDB) and simulate real user journeys.

## Validation Commands
- `make test-e2e`
- `make test-all`

## Prerequisites
- Plans 00 (Go upgrade) and 01 (Auth) should be completed first,
  as E2E tests will use authentication

---

### Task 1: Set up E2E test infrastructure
- [ ] Create `test/integration/e2e/suite_test.go` with test setup
- [ ] Extend `shared.IntegrationSuite` with auth helpers (login, get token)
- [ ] Add helper to create seeded data (admin user, test organization, categories)
- [ ] Mark completed

### Task 2: Ticket lifecycle workflow
- [ ] Test: Admin creates user (Agent) -> Agent logs in -> Agent creates ticket ->
      Agent changes status to in_progress -> Agent resolves ticket -> Admin closes ticket
- [ ] Verify status transitions follow domain rules
- [ ] Verify timestamps are set correctly (resolvedAt, closedAt)
- [ ] Mark completed

### Task 3: User management workflow
- [ ] Test: Admin creates users with different roles -> Users login ->
      Users see only their own tickets -> Admin changes user role ->
      Verify new permissions take effect
- [ ] Mark completed

### Task 4: Organization workflow
- [ ] Test: Admin creates organization -> Admin adds users ->
      Create tickets under organization -> Filter tickets by organization ->
      Verify organization users see org tickets
- [ ] Mark completed

### Task 5: Category and ticket classification
- [ ] Test: Create category tree -> Create tickets in categories ->
      Filter tickets by category -> Move category (change parent) ->
      Verify ticket associations remain correct
- [ ] Mark completed

### Task 6: Error scenarios
- [ ] Test: Invalid status transitions return proper errors
- [ ] Test: Duplicate email on user creation returns 409
- [ ] Test: Circular parent references in categories/organizations return errors
- [ ] Test: Accessing nonexistent resources returns 404
- [ ] Mark completed
