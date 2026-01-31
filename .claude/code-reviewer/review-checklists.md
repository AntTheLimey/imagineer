# Review Checklists

This document provides checklists for reviewing different types of changes.

## New Feature Checklist

### Before Review

- [ ] Feature described in PR/commit message
- [ ] Tests pass locally
- [ ] No linter warnings

### Code Quality

- [ ] Functions are appropriately sized (< 50 lines)
- [ ] Nesting depth is reasonable (< 4 levels)
- [ ] Error handling is complete
- [ ] No unused code or imports
- [ ] Follows existing patterns

### Testing

- [ ] Unit tests added
- [ ] Edge cases covered
- [ ] Test names are descriptive
- [ ] Mocks are appropriate

### Documentation

- [ ] Public APIs documented
- [ ] Complex logic explained
- [ ] README updated if needed

### Security

- [ ] Input validation present
- [ ] No SQL injection vectors
- [ ] Sensitive data protected
- [ ] GM notes not exposed

### Imagineer-Specific

- [ ] Entity types validated
- [ ] Canon confidence handled
- [ ] Duplicate name check present
- [ ] Game system schema respected

## Bug Fix Checklist

### Understanding

- [ ] Root cause identified
- [ ] Fix addresses root cause (not symptom)
- [ ] No regression risks

### Code Changes

- [ ] Minimal changes to fix issue
- [ ] No unrelated changes included
- [ ] Error handling improved

### Testing

- [ ] Test reproduces the bug
- [ ] Test passes with fix
- [ ] Related tests still pass

### Documentation

- [ ] Code comments updated if needed
- [ ] CHANGELOG entry added

## Refactoring Checklist

### Planning

- [ ] Refactoring scope is clear
- [ ] Behavior should be unchanged
- [ ] Tests exist before refactoring

### Execution

- [ ] Small, incremental changes
- [ ] Tests pass after each change
- [ ] No functional changes mixed in

### Verification

- [ ] All existing tests pass
- [ ] No new warnings introduced
- [ ] Performance not degraded

## Importer Change Checklist

### Interface Compliance

- [ ] Implements Importer interface
- [ ] Returns proper ImportResult
- [ ] Handles ImportOptions correctly

### Entity Extraction

- [ ] Entity types detected correctly
- [ ] Names extracted properly
- [ ] Descriptions preserved
- [ ] Tags captured
- [ ] Source document tracked

### Relationship Extraction

- [ ] Relationship patterns recognized
- [ ] Source/target entities identified
- [ ] Bidirectional flag set correctly

### Event Extraction

- [ ] Date patterns recognized
- [ ] Date precision determined
- [ ] Related entities linked

### Error Handling

- [ ] Parse errors captured in result
- [ ] Warnings for partial failures
- [ ] Graceful handling of malformed input

### Testing

- [ ] Sample files tested
- [ ] Edge cases covered
- [ ] Malformed input handled

## API Endpoint Checklist

### Request Handling

- [ ] Input validation complete
- [ ] Required fields checked
- [ ] Types validated
- [ ] Size limits enforced

### Authorization

- [ ] Authentication required
- [ ] Campaign access verified
- [ ] GM vs player access checked

### Response

- [ ] Success response format correct
- [ ] Error responses informative
- [ ] No sensitive data leaked
- [ ] Proper HTTP status codes

### Testing

- [ ] Happy path tested
- [ ] Error cases tested
- [ ] Authorization tested

## Database Migration Checklist

### Migration Design

- [ ] Schema changes are additive if possible
- [ ] Rollback plan exists
- [ ] Data migration tested

### SQL Quality

- [ ] Indexes added for foreign keys
- [ ] Indexes added for frequently queried columns
- [ ] COMMENT ON added for tables/columns
- [ ] Transaction used for complex changes

### Compatibility

- [ ] Existing data handled
- [ ] Application code compatible
- [ ] No breaking changes to API

### Testing

- [ ] Migration runs successfully
- [ ] Data integrity maintained
- [ ] Rollback works

## React Component Checklist

### Component Design

- [ ] Single responsibility
- [ ] Props are typed
- [ ] Default props provided where appropriate

### State Management

- [ ] State is minimal
- [ ] Derived state computed, not stored
- [ ] Hooks used correctly

### Performance

- [ ] Memoization where needed
- [ ] No unnecessary re-renders
- [ ] Lists have stable keys

### Accessibility

- [ ] Semantic HTML used
- [ ] ARIA labels present
- [ ] Keyboard navigation works

### Testing

- [ ] Component renders
- [ ] User interactions tested
- [ ] Edge cases handled
