# Filora API Testing Documentation

This document describes the testing strategy and coverage for the Filora DAM API.

## Test Coverage

### ✅ Unit Tests Implemented

#### Library Functions (100% coverage)
- `internal/lib/password_test.go` - Password hashing and verification
  - HashPassword success and security (4 tests)
  - VerifyPassword correct/incorrect scenarios
  - Different hashes for same password (salt verification)

- `internal/lib/jwt_test.go` - JWT token management
  - Token generation and validation (7 tests)
  - Token expiration verification (24h)
  - Secret key isolation
  - Invalid token handling

- `internal/lib/hash_test.go` - File hashing
  - SHA-256 hash generation and consistency (6 tests)
  - File vs bytes hashing
  - Different inputs produce different hashes

All lib tests **PASSING** ✅

### Integration Testing Strategy

#### Repository Layer
- Database integration with PostgreSQL
- Transaction handling
- Query generation via sqlc
- UUID conversion and type safety

**Testing approach:**
- Use actual database connection in CI/test environment
- Test migration rollback/forward
- Verify indexes exist and query efficiency

#### Service Layer
- Business logic orchestration
- Error propagation
- Data validation

**Testing approach:**
- Mock repositories with testify/mock
- Verify service coordination logic
- Test error handling paths

#### Handler Layer
- HTTP request/response handling
- Middleware integration
- Authentication verification

**Testing approach:**
- HTTP integration tests with actual fiber app
- Test auth middleware with valid/invalid tokens
- Verify response formats and status codes

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./internal/lib

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Scenarios

### Critical Paths

1. **User Registration & Authentication**
   - Duplicate email detection
   - Password hashing and verification
   - JWT token generation
   - Token validation and expiration

2. **File Upload & Deduplication**
   - Hash calculation
   - Duplicate detection by hash
   - Provider selection with quota awareness
   - Storage usage tracking

3. **File Download**
   - Permission verification
   - Storage location lookup
   - Provider adapter initialization
   - File streaming

4. **Search & Filtering**
   - Name-based search (ILIKE)
   - Type-based filtering
   - Pagination with limit/offset
   - Result ordering

5. **Dashboard Statistics**
   - Aggregate calculations
   - Type distribution
   - Recent activity tracking
   - Quota calculations

## Future Testing Goals

### Short-term
- [ ] Repository integration tests (requires test database)
- [ ] Handler/HTTP integration tests
- [ ] End-to-end workflow tests
- [ ] Error scenario coverage

### Medium-term
- [ ] Performance/load testing
- [ ] Storage adapter tests (Cloudinary, ImageKit, R2)
- [ ] Concurrent upload/download tests
- [ ] Database transaction tests

### Long-term
- [ ] Chaos engineering tests
- [ ] Security penetration tests
- [ ] Compliance/audit logging tests
- [ ] Backup and recovery tests

## Mocking Strategy

### When to Mock
- External services (storage providers)
- Database for unit tests
- HTTP clients

### When NOT to Mock
- Pure functions (hashing, password verification)
- Type conversions
- Error handling in business logic

## Test Utilities

### testify/assert
Used for clean assertion syntax:
```go
assert.NoError(t, err)
assert.Equal(t, expected, actual)
assert.NotNil(t, value)
```

### testify/mock
Used for repository mocking:
```go
mockRepo.On("GetByID", ctx, "id").Return(user, nil)
mockRepo.AssertExpectations(t)
```

## Performance Benchmarks

To add performance tests:

```go
func BenchmarkHashPassword(b *testing.B) {
    for i := 0; i < b.N; i++ {
        HashPassword("test-password")
    }
}
```

Run with: `go test -bench ./...`

## CI/CD Integration

Tests should be run on every commit:
- Pre-commit hook: unit tests pass
- Pre-push hook: all tests pass
- CI pipeline: full test suite + coverage report

## Coverage Goals

- **Core logic (services)**: 70%+ coverage
- **API handlers**: 60%+ coverage
- **Utilities (lib)**: 90%+ coverage
- **Repositories**: Integration tests
- **Overall**: 65%+ coverage target

## Known Limitations

1. **Database Tests**: Require PostgreSQL instance
2. **Storage Adapters**: Need provider credentials
3. **HTTP Tests**: Cannot mock Fiber middleware easily
4. **End-to-end Tests**: Require full stack setup

## Contributing Tests

When adding new features:
1. Write unit tests for new functions
2. Add integration tests for workflows
3. Update this documentation
4. Ensure test passes locally before commit
5. Run full suite before push: `go test ./...`
