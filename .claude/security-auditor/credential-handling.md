# Credential Handling

This document describes secure credential and secret management.

## Secret Categories

### Application Secrets

| Secret | Storage | Usage |
|--------|---------|-------|
| Database password | Environment variable | Connection string |
| Session secret | Environment variable | Cookie signing |
| API keys | Environment variable | External services |

### User Credentials

| Credential | Storage | Handling |
|------------|---------|----------|
| Passwords | Database (hashed) | bcrypt hash |
| Session tokens | Database + cookie | Cryptographic random |

## Environment Variables

### Required Variables

```bash
# Database
IMAGINEER_DB_HOST=localhost
IMAGINEER_DB_PORT=5432
IMAGINEER_DB_NAME=imagineer
IMAGINEER_DB_USER=imagineer
IMAGINEER_DB_PASSWORD=<secret>

# Session
IMAGINEER_SESSION_SECRET=<32+ random bytes>

# Server
IMAGINEER_PORT=8080
```

### Loading Variables

```go
func loadConfig() (*Config, error) {
    password := os.Getenv("IMAGINEER_DB_PASSWORD")
    if password == "" {
        return nil, errors.New("IMAGINEER_DB_PASSWORD not set")
    }
    // Never log the password
    return &Config{
        DBPassword: password,
    }, nil
}
```

## Password Handling

### Hashing Passwords

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword(
        []byte(password),
        bcrypt.DefaultCost,
    )
    return string(bytes), err
}
```

### Verifying Passwords

```go
func verifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword(
        []byte(hash),
        []byte(password),
    )
    return err == nil
}
```

### Password Requirements

| Requirement | Value |
|-------------|-------|
| Minimum length | 8 characters |
| Maximum length | 72 characters (bcrypt limit) |

## Session Token Generation

### Creating Tokens

```go
import "crypto/rand"

func generateToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}
```

### Token Storage

```go
// Store hashed token in database
hashedToken := sha256.Sum256([]byte(token))
db.StoreSession(userID, hashedToken[:])

// Send original token to client (once)
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    token,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
})
```

## Connection Strings

### Building Connection Strings

```go
func buildConnString(cfg *Config) string {
    // Use structured connection, not string interpolation
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
        cfg.DBHost,
        cfg.DBPort,
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBName,
    )
}
```

### Connection String Security

- Never log connection strings (contain password)
- Use SSL/TLS (`sslmode=require`)
- Rotate passwords periodically

## Logging Guidelines

### Never Log

```go
// NEVER log these
log.Info("password", password)
log.Info("token", sessionToken)
log.Info("connection", connString)
log.Info("api_key", apiKey)
```

### Safe Logging

```go
// Safe to log
log.Info("user_login", "username", username)
log.Info("session_created", "user_id", userID)
log.Info("database_connected", "host", cfg.DBHost)
```

## Error Messages

### Internal Errors

```go
// Log detailed error internally
log.Error("authentication failed",
    "username", username,
    "error", err,
    "ip", remoteAddr,
)
```

### External Errors

```go
// Return generic error to client
return &APIError{
    Code:    "AUTH_FAILED",
    Message: "Invalid username or password",
}
```

## Development vs Production

### Development

```bash
# .env file (in .gitignore)
IMAGINEER_DB_PASSWORD=devpassword
IMAGINEER_SESSION_SECRET=devsecret
```

### Production

```bash
# From secret manager or environment
export IMAGINEER_DB_PASSWORD=$(vault read ...)
export IMAGINEER_SESSION_SECRET=$(vault read ...)
```

## Credential Rotation

### Database Password Rotation

1. Create new password in database
2. Update environment variable
3. Restart application
4. Remove old password

### Session Secret Rotation

1. Add new secret (support both old and new)
2. Wait for old sessions to expire
3. Remove old secret support

## Security Checklist

### Before Deployment

- [ ] All secrets in environment variables
- [ ] No hardcoded credentials
- [ ] .env in .gitignore
- [ ] SSL enabled for database
- [ ] Passwords properly hashed
- [ ] Session tokens cryptographically random
- [ ] Connection strings not logged
