# Development Plans

## Execution Order

Plans are numbered and should be executed in order. Some have dependencies on earlier plans.

| # | Plan | Description | Dependencies |
|---|------|-------------|--------------|
| 00 | [Upgrade Go 1.26](00-upgrade-go-1.26.md) | Upgrade from Go 1.24 to 1.26 | None |
| 01 | [Authentication & Authorization](01-authentication-authorization.md) | JWT auth, login endpoint, role-based access | 00 |
| 02 | [Request Validation](02-request-validation.md) | OpenAPI schema validation middleware | 00 |
| 03 | [CORS & Rate Limiting](03-cors-rate-limiting.md) | Frontend support and abuse protection | 01 |
| 04 | [Health Check](04-health-check.md) | Liveness/readiness probes with DB checks | 00 |
| 05 | [E2E Tests](05-e2e-tests.md) | Full workflow end-to-end tests | 01 |
| 06 | [Update Documentation](06-update-documentation.md) | Sync all docs after implementation | All |

## Status

- [ ] 00 - Upgrade Go 1.26 (in progress; remaining validation checkboxes open)
- [x] 01 - Authentication & Authorization
- [x] 02 - Request Validation
- [ ] 03 - CORS & Rate Limiting
- [ ] 04 - Health Check
- [ ] 05 - E2E Tests
- [ ] 06 - Update Documentation (in progress)
