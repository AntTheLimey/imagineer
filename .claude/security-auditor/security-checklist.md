# Security Checklist

Component-specific security checklists for Imagineer.

## API Handler Checklist

### Authentication

- [ ] Handler requires authentication
- [ ] Authentication verified before processing
- [ ] Failed auth attempts logged
- [ ] No timing leaks in auth checks

### Authorization

- [ ] Campaign ownership verified
- [ ] GM vs player role checked
- [ ] GM notes filtered for players
- [ ] Resource access matches user

### Input Validation

- [ ] All parameters validated
- [ ] Type checking performed
- [ ] Length limits enforced
- [ ] Format validation for IDs, dates, URLs

### Output Security

- [ ] Sensitive data filtered
- [ ] Error messages are generic
- [ ] No stack traces to client
- [ ] Proper HTTP status codes

## Database Query Checklist

### Query Safety

- [ ] Parameterized queries only
- [ ] No string concatenation in SQL
- [ ] No user input in table/column names
- [ ] Query results limited

### Connection Security

- [ ] SSL/TLS enabled
- [ ] Credentials from environment
- [ ] Connection string not logged
- [ ] Minimal database privileges

## Import Handler Checklist

### File Upload (Evernote)

- [ ] File size limit enforced
- [ ] File type validated (.enex)
- [ ] XML parser protected against XXE
- [ ] Parsing depth limited

### URL Fetch (Google Docs)

- [ ] URL format validated
- [ ] Only google.com domains
- [ ] Response size limited
- [ ] Timeout enforced

### Content Processing

- [ ] Extracted content sanitized
- [ ] Entity names validated
- [ ] No script execution
- [ ] Safe JSONB construction

## React Component Checklist

### Data Display

- [ ] No dangerouslySetInnerHTML
- [ ] User content escaped
- [ ] URLs validated before linking
- [ ] Images from trusted sources

### Form Handling

- [ ] Client-side validation
- [ ] No sensitive data in URL
- [ ] Form tokens if needed
- [ ] Secure file upload

### State Management

- [ ] No secrets in client state
- [ ] Session tokens in httpOnly cookies
- [ ] Sensitive data cleared on logout

## Session Management Checklist

### Session Creation

- [ ] Cryptographic random token
- [ ] Token hashed in database
- [ ] Secure cookie flags set
- [ ] Expiration configured

### Session Validation

- [ ] Constant-time comparison
- [ ] Expiration checked
- [ ] User still valid
- [ ] IP/device binding (optional)

### Session Termination

- [ ] Logout invalidates session
- [ ] Expired sessions cleaned up
- [ ] All user sessions revokable

## Entity Handling Checklist

### Creation

- [ ] Entity type validated
- [ ] Campaign access verified
- [ ] Duplicate name check performed
- [ ] JSONB attributes validated

### Retrieval

- [ ] Campaign ownership checked
- [ ] GM notes filtered for players
- [ ] Related data filtered
- [ ] Query optimized

### Update

- [ ] Authorization verified
- [ ] All fields validated
- [ ] Canon conflict checked
- [ ] Audit trail maintained

### Deletion

- [ ] Authorization verified
- [ ] Related data handled
- [ ] Soft delete preferred
- [ ] Audit trail maintained

## Configuration Checklist

### Environment

- [ ] Secrets in environment variables
- [ ] No default passwords
- [ ] .env in .gitignore
- [ ] Production config separate

### Logging

- [ ] No sensitive data logged
- [ ] Security events logged
- [ ] Log injection prevented
- [ ] Logs protected

### Error Handling

- [ ] Generic errors to users
- [ ] Detailed errors to logs
- [ ] No stack traces exposed
- [ ] Error codes documented

## Pre-Deployment Checklist

### Code Review

- [ ] Security-focused review completed
- [ ] All inputs validated
- [ ] All outputs filtered
- [ ] No hardcoded secrets

### Testing

- [ ] Security tests pass
- [ ] Injection tests performed
- [ ] Authorization tests pass
- [ ] Error handling tested

### Configuration

- [ ] HTTPS enforced
- [ ] Security headers set
- [ ] Rate limiting enabled
- [ ] Monitoring configured

## Incident Response Checklist

### If GM Notes Exposed

1. [ ] Identify affected campaigns
2. [ ] Notify affected users
3. [ ] Audit access logs
4. [ ] Fix vulnerability
5. [ ] Review similar code
6. [ ] Update tests

### If Unauthorized Access

1. [ ] Revoke affected sessions
2. [ ] Reset affected passwords
3. [ ] Audit access logs
4. [ ] Fix vulnerability
5. [ ] Notify affected users
6. [ ] Review authorization code

### If SQL Injection

1. [ ] Take endpoint offline
2. [ ] Audit for data access
3. [ ] Fix vulnerability
4. [ ] Review all queries
5. [ ] Add parameterization tests
6. [ ] Restore service
