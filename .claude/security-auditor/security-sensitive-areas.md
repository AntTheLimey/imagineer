# Security-Sensitive Areas

This document identifies code locations requiring security review.

## Critical Risk Areas

### Entity Access Control

**Risk Level: CRITICAL**

| Location | Security Concern |
|----------|------------------|
| `/internal/api/entities.go` | GM note exposure, authorization |
| `/internal/models/entity.go` | Field filtering for player views |

**Review checklist:**

- [ ] GM notes filtered for player requests
- [ ] Campaign ownership verified
- [ ] Entity type validated
- [ ] Input sanitized

### Import Processing

**Risk Level: HIGH**

| Location | Security Concern |
|----------|------------------|
| `/internal/importers/evernote/` | XML parsing, entity injection |
| `/internal/importers/googledocs/` | URL validation, content injection |

**Review checklist:**

- [ ] XML parsing doesn't allow XXE
- [ ] URLs validated before fetching
- [ ] Extracted content sanitized
- [ ] File size limits enforced

### Database Operations

**Risk Level: HIGH**

| Location | Security Concern |
|----------|------------------|
| `/internal/database/` | SQL injection, credential handling |
| `/migrations/` | Schema security |

**Review checklist:**

- [ ] All queries parameterized
- [ ] Connection credentials from environment
- [ ] No credentials in logs
- [ ] Minimal database privileges

## High Risk Areas

### API Handlers

**Risk Level: HIGH**

| Location | Security Concern |
|----------|------------------|
| `/internal/api/` | Input validation, authorization |
| `/cmd/server/` | Server configuration |

**Review checklist:**

- [ ] All input validated
- [ ] Authorization on every endpoint
- [ ] Rate limiting considered
- [ ] Error messages don't leak info

### Session Management

**Risk Level: HIGH**

| Location | Security Concern |
|----------|------------------|
| `/internal/api/sessions.go` | Session data exposure |

**Review checklist:**

- [ ] Prep notes protected (GM only)
- [ ] Session ownership verified
- [ ] Discovery data filtered by role

## Medium Risk Areas

### Client Input Handling

**Risk Level: MEDIUM**

| Location | Security Concern |
|----------|------------------|
| `/client/src/pages/` | XSS, input validation |
| `/client/src/components/` | User data display |

**Review checklist:**

- [ ] No dangerouslySetInnerHTML with user data
- [ ] Input validated before submission
- [ ] API tokens stored securely

### Configuration

**Risk Level: MEDIUM**

| Location | Security Concern |
|----------|------------------|
| `/config/` | Secret management |
| `.env` | Credential storage |

**Review checklist:**

- [ ] Secrets in environment variables
- [ ] No default credentials
- [ ] .env in .gitignore

## Code Patterns to Flag

### Always Flag

```go
// SQL string concatenation - ALWAYS VULNERABLE
query := "SELECT * FROM entities WHERE id = " + entityID

// GM notes in player response
return entity  // Without filtering

// Weak random for security
import "math/rand"  // Should be crypto/rand

// Hardcoded credentials
password := "secret123"
```

### Review Carefully

```go
// Error messages with internal details
return fmt.Errorf("entity %s not found in campaign %s", id, campaignID)

// File operations with user input
os.Open(userPath)  // Path traversal risk

// HTTP requests with user URLs
http.Get(userURL)  // SSRF risk
```

### Acceptable Patterns

```go
// Parameterized queries - SAFE
row := db.QueryRow(ctx, "SELECT * FROM entities WHERE id = $1", id)

// Filtered entity response - SAFE
return filterForPlayer(entity)

// Crypto random - SAFE
token := make([]byte, 32)
crypto_rand.Read(token)
```

## Trust Boundaries

```
UNTRUSTED                    TRUST BOUNDARY                    TRUSTED
─────────────────────────────────────────────────────────────────────────
User input                        │
Import files       ────────────▶  │ ────────────▶  Validated data
API requests                      │                Server logic
                                  │
Player requests    ────────────▶  │ ────────────▶  Filtered entities
                                  │                (no GM notes)
```

## Data Classification

| Classification | Examples | Handling |
|----------------|----------|----------|
| **Secret** | Session tokens | Never logged, encrypted |
| **GM-Only** | GM notes, prep notes | Never sent to players |
| **Campaign-Private** | All campaign data | Access control required |
| **Public** | Game system schemas | No restrictions |
