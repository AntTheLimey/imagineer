# Testing Expert Knowledge Base

This directory contains testing guidance for Imagineer.

## Documents

- `testing-overview.md` - Testing strategy
- `unit-testing.md` - Unit test patterns
- `integration-testing.md` - Integration tests
- `database-testing.md` - Database test patterns

## Quick Reference

### Test Commands

```bash
# All tests
make test-all

# Go tests
go test ./...

# React tests
npm test

# Coverage
make coverage
```

### Coverage Goals

| Component | Target |
|-----------|--------|
| Overall | >80% |
| Database operations | >90% |
| Entity management | >90% |
| Security functions | 100% |

### Test File Locations

| Type | Location |
|------|----------|
| Go unit tests | `*_test.go` (co-located) |
| React tests | `/client/src/**/*.test.tsx` |
| Integration | `/tests/integration/` |

Last Updated: 2026-01-30
