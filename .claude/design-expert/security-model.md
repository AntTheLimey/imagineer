# Security Model

This document describes the security architecture and access control model
for Imagineer.

## Security Goals

### Primary Goals

1. **Data Confidentiality**: Keeper notes never exposed to players
2. **Data Integrity**: Canon conflicts detected, not silently resolved
3. **Access Control**: Users only access their own campaigns
4. **Input Validation**: All user input validated before processing

### Threat Model

**In Scope:**

- Unauthorized access to campaigns
- Keeper note exposure to players
- SQL injection attacks
- Cross-site scripting (XSS)
- Data tampering

**Out of Scope (for initial version):**

- Multi-tenant isolation
- Federated authentication
- Advanced persistent threats

## Access Control Architecture

### User Roles

| Role | Description | Capabilities |
|------|-------------|--------------|
| Keeper | Game Master | Full access to campaign |
| Player | Campaign participant | Read public entities, no keeper notes |
| Guest | Unauthenticated | No access |

### Resource Ownership

```
User → owns → Campaign → contains → Entity
                      → contains → Session
                      → contains → Relationship
                      → contains → TimelineEvent
                      → contains → CanonConflict
```

### Authorization Rules

1. **Campaign Access**: Only owner can access campaign
2. **Entity Visibility**: Players see entities minus keeper notes
3. **Session Access**: Only keeper can see prep notes
4. **Conflict Resolution**: Only keeper can resolve conflicts

## Keeper Note Protection

### Critical Requirement

Keeper notes (GM-only content) must NEVER be exposed to players.

### Implementation

```go
// Entity fields with keeper-only content
type Entity struct {
    // Public fields
    ID          string
    Name        string
    Description string
    // ...

    // KEEPER-ONLY FIELDS
    KeeperNotes       string  // Never send to players
    DiscoveredSession string  // May reveal plot
}

// Filter for player access
func FilterForPlayer(entity *Entity) *Entity {
    return &Entity{
        ID:          entity.ID,
        Name:        entity.Name,
        Description: entity.Description,
        // KeeperNotes omitted
        // DiscoveredSession omitted
    }
}
```

### API Enforcement

All API endpoints that return entities must:

1. Check if requester is keeper or player
2. Filter keeper-only fields for players
3. Never include keeper notes in list responses to players

## Input Validation

### Validation Layers

```
Client Input → Client Validation → Server Validation → Database Constraints
```

### Server Validation Requirements

All input must be validated for:

- **Type correctness**: Match expected types
- **Length limits**: Enforce maximum lengths
- **Format validation**: Match expected patterns
- **Business rules**: Valid entity types, etc.

### SQL Injection Prevention

```go
// ALWAYS use parameterized queries
row := db.QueryRow(ctx, "SELECT * FROM entities WHERE id = $1", entityID)

// NEVER concatenate user input
query := "SELECT * FROM entities WHERE id = " + entityID  // VULNERABLE
```

### XSS Prevention

```typescript
// NEVER use dangerouslySetInnerHTML with user data
<div dangerouslySetInnerHTML={{ __html: userContent }} />  // VULNERABLE

// ALWAYS sanitize or use React's built-in escaping
<div>{userContent}</div>  // Safe - React escapes
```

## Authentication

### Current Approach (Development)

For initial development, authentication is simplified:

- Session-based authentication
- Secure cookies
- HTTPS required in production

### Password Handling

```go
// Hash passwords with bcrypt
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(password),
    bcrypt.DefaultCost,
)

// Compare passwords
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
```

### Session Management

- Sessions expire after inactivity
- Sessions invalidated on logout
- Session tokens are cryptographically random

## Database Security

### Connection Security

- Use SSL/TLS for database connections
- Credentials in environment variables
- Connection string never logged

### Query Security

- Parameterized queries only
- No dynamic SQL construction
- Minimal privileges for application user

### Data at Rest

- JSONB data is not encrypted (future consideration)
- Backup encryption recommended

## Secure Coding Practices

### Error Handling

```go
// DON'T expose internal details
return fmt.Errorf("database error: %v", err)  // May leak info

// DO use generic messages for users
log.Error("database error", "err", err)
return errors.New("an error occurred")
```

### Logging

```go
// NEVER log sensitive data
log.Info("user login", "password", password)  // WRONG

// DO log safely
log.Info("user login", "username", username)
```

### Random Number Generation

```go
// DON'T use math/rand for security
token := make([]byte, 32)
rand.Read(token)  // Predictable!

// DO use crypto/rand
token := make([]byte, 32)
crypto_rand.Read(token)  // Secure
```

## Security Checklist

### New Feature Checklist

- [ ] Input validated on server
- [ ] Authorization checked
- [ ] Keeper notes protected
- [ ] No SQL injection vectors
- [ ] No XSS vectors
- [ ] Errors don't leak info
- [ ] Sensitive data not logged

### API Endpoint Checklist

- [ ] Authentication required
- [ ] Authorization verified
- [ ] Input validated
- [ ] Output filtered by role
- [ ] Rate limiting considered

### Database Query Checklist

- [ ] Parameterized query used
- [ ] Result set limited
- [ ] Sensitive columns filtered
- [ ] Campaign access verified

## Incident Response

### If Keeper Notes Exposed

1. Identify affected campaigns
2. Notify affected users
3. Audit access logs
4. Fix vulnerability
5. Review similar code

### If SQL Injection Found

1. Take affected endpoint offline
2. Fix vulnerability
3. Audit for data exfiltration
4. Review all queries
5. Add automated testing

## Future Considerations

### Multi-Tenant Support

When adding multi-tenant support:

- Row-level security in PostgreSQL
- Tenant isolation at query level
- Separate encryption keys per tenant

### Player Access

When adding player portals:

- Strict field filtering
- Session-based visibility
- Discovery tracking
