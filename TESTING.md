# Testing Guide for Farmers Module

This document provides comprehensive information about testing the farmers-module with TestContainers, contract tests, and security-enhanced mocks.

## Table of Contents

- [Overview](#overview)
- [Test Categories](#test-categories)
- [Running Tests](#running-tests)
- [TestContainers Setup](#testcontainers-setup)
- [Contract Testing](#contract-testing)
- [Security Testing](#security-testing)
- [Mock Factory Usage](#mock-factory-usage)
- [Best Practices](#best-practices)

## Overview

The farmers-module uses a comprehensive testing strategy that includes:

1. **Unit Tests**: Fast, isolated tests using mocks
2. **Integration Tests**: Tests with real PostgreSQL+PostGIS using TestContainers
3. **Contract Tests**: Ensures mocks behave like real services
4. **Security Tests**: Validates security controls and attack resistance

## Test Categories

### Unit Tests

Fast tests that use mocks and don't require external dependencies.

```bash
# Run only unit tests (skip integration tests)
make test-short
```

Unit tests should:
- Execute in milliseconds
- Use mocks for all external dependencies
- Test business logic in isolation
- Not require Docker or databases

### Integration Tests

Tests that use real PostgreSQL with PostGIS via TestContainers.

```bash
# Run integration tests (requires Docker)
make test-integration
```

Integration tests:
- Use real PostgreSQL+PostGIS containers
- Test spatial operations and queries
- Validate database migrations
- Test complete workflows end-to-end

**Prerequisites**: Docker must be running

### Contract Tests

Validates that mocks behave exactly like real service implementations.

```bash
# Run contract tests
make test-contract
```

Contract tests ensure:
- Mock behavior matches real service behavior
- No behavioral drift over time
- Consistent API contracts
- Proper error handling

### Security Tests

Validates security controls and tests attack resistance.

```bash
# Run security tests
make test-security
```

Security tests cover:
- JWT signature validation
- Rate limiting enforcement
- Audit logging
- Common attack vectors (SQL injection, XSS, etc.)
- Permission boundary enforcement

## Running Tests

### All Tests

Run all test categories:

```bash
make test
```

### Specific Test Categories

```bash
# Unit tests only
make test-short

# Integration tests only
make test-integration

# Contract tests only
make test-contract

# Security tests only
make test-security

# Benchmark tests
make test-benchmark
```

### Coverage Reports

Generate test coverage reports:

```bash
make test-coverage
```

This creates:
- `coverage.out` - Coverage data file
- `coverage.html` - HTML coverage report

Open `coverage.html` in a browser to view detailed coverage.

### Running Specific Tests

```bash
# Run a specific test file
go test ./internal/services -v -run TestMockFactory

# Run tests matching a pattern
go test ./... -v -run "TestAAA.*"

# Run with verbose output
go test ./... -v

# Run with race detection
go test ./... -race
```

## TestContainers Setup

### Overview

TestContainers provides real PostgreSQL+PostGIS databases for integration testing. Each test gets an isolated container.

### Basic Usage

```go
import "github.com/Kisanlink/farmers-module/internal/testutils"

func TestWithPostgreSQL(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup PostgreSQL container with PostGIS
    pgContainer := testutils.SetupPostgreSQLContainer(t)

    // Setup database with migrations
    db := pgContainer.SetupTestDB(t, &MyModel{})

    // Use the database
    // ...

    // Cleanup is automatic via t.Cleanup()
}
```

### Parallel Test Execution

For better performance, tests can run in parallel:

```go
func TestParallelExample(t *testing.T) {
    db, cleanup := testutils.SetupParallelTestDB(t, &MyModel{})
    defer cleanup()

    // Each parallel test gets its own isolated database
    // ...
}
```

### PostGIS Operations

The helper provides PostGIS-specific utilities:

```go
// Validate PostGIS installation
pgContainer.ValidatePostGISInstallation(t, db)

// Create spatial indexes
pgContainer.CreateSpatialIndex(t, db, "farms", "geometry")

// Test database connection
pgContainer.TestDatabaseConnection(t, db)

// Reset database between tests
pgContainer.ResetDatabase(t, db)
```

### Configuration

Default configuration:
- **Image**: `postgis/postgis:16-3.5`
- **Database**: `farmers_test`
- **Username**: `test`
- **Password**: `test`
- **Startup Timeout**: 60 seconds

## Contract Testing

### Purpose

Contract tests ensure mock implementations behave exactly like real services, preventing behavioral drift.

### Running Contract Tests

All mock implementations must pass the contract tests:

```go
// Run all AAA service contract tests
func TestMyMockImplementation(t *testing.T) {
    mock := createMyMock()
    RunContractTests(t, mock, "MyMockImplementation")
}
```

### Behavior Drift Detection

The behavior drift detector identifies when mocks diverge from expected behavior:

```go
detector := NewBehaviorDriftDetector()

// Run tests and record violations
// ...

// Print report
detector.PrintReport(t)
```

Violations are categorized by severity:
- **Critical**: Security issues, must fix immediately
- **High**: Major functionality differences
- **Medium**: Minor behavioral differences
- **Low**: Edge case handling differences

### Adding New Contract Tests

To add a new contract test:

```go
var NewServiceContractTests = []ContractTest{
    {
        name: "test_name",
        description: "What this test validates",
        testFunc: func(t *testing.T, service ServiceInterface) {
            // Test implementation
        },
    },
}
```

## Security Testing

### Security-Enhanced Mocks

Security-enhanced mocks provide JWT validation, rate limiting, and audit logging:

```go
// Create security-enhanced mock
mock, err := NewSecurityEnhancedMockAAA(true)
require.NoError(t, err)

// Generate test token
token, err := mock.GenerateTestToken("user123", "org456", []string{"admin"})

// Validate token (includes JWT signature validation)
userInfo, err := mock.ValidateToken(ctx, token)

// Check audit logs
events := mock.GetAuditEvents()
```

### JWT Validation

The JWT validator uses RSA signatures for realistic token validation:

```go
validator := mock.GetJWTValidator()

// Generate token
token, _ := validator.GenerateToken("user123", "org456", []string{"admin"})

// Validate token (checks signature, expiry, issuer)
claims, err := validator.ValidateToken(token)
```

### Rate Limiting

Simulate rate limiting for testing:

```go
rateLimiter := mock.GetRateLimiter()

// Set custom limit
rateLimiter.SetLimit("user123", 10)

// Test rate limiting
for i := 0; i < 100; i++ {
    allowed, err := rateLimiter.Allow("user123")
    if !allowed {
        // Rate limit exceeded
    }
}

// Reset limits
rateLimiter.Reset("user123")
```

### Audit Logging

All security operations are logged:

```go
auditLogger := mock.auditLog

// Get all events
events := auditLogger.GetEvents()

// Filter by user
userEvents := auditLogger.GetEventsByUser("user123")

// Filter by type
permissionEvents := auditLogger.GetEventsByType("permission_check")

// Clear logs
auditLogger.Clear()
```

### Attack Scenario Testing

Test common attack vectors:

```go
func TestAttackScenarios(t *testing.T) {
    mock, _ := NewSecurityEnhancedMockAAA(true)

    // SQL injection attempt
    mock.CheckPermission(ctx, "admin' OR '1'='1", ...)

    // XSS attempt
    mock.CheckPermission(ctx, "<script>alert('xss')</script>", ...)

    // Brute force rate limiting
    for i := 0; i < 1000; i++ {
        mock.CheckPermission(ctx, "attacker", ...)
    }

    // Token forgery
    mock.ValidateToken(ctx, "forged.jwt.token")
}
```

## Mock Factory Usage

The mock factory provides centralized mock creation with validation and presets.

### Creating Mocks

```go
// Create factory with security defaults
factory := NewMockFactory(PermissionModeDenyAll)

// Create AAA service mock
aaaService := factory.NewAAAServiceMock()

// Create cache mock
cache := factory.NewCacheMock()

// Create event emitter mock
eventEmitter := factory.NewEventEmitterMock()

// Create database mock
database := factory.NewDatabaseMock()
```

### Using Presets

Presets provide pre-configured permission sets for common scenarios:

```go
factory := NewMockFactory(PermissionModeDenyAll)

// Admin preset - full permissions
adminMock := factory.NewAAAServiceMockWithPreset(PresetAdmin, "admin123", "org123")

// Farmer preset - limited self-service permissions
farmerMock := factory.NewAAAServiceMockWithPreset(PresetFarmer, "farmer123", "org123")

// KisanSathi preset - farmer management permissions
ksMock := factory.NewAAAServiceMockWithPreset(PresetKisanSathi, "ks123", "org123")

// FPO Manager preset - organization management permissions
fpoMock := factory.NewAAAServiceMockWithPreset(PresetFPOManager, "mgr123", "org123")

// Read-only preset - no write permissions
readonlyMock := factory.NewAAAServiceMockWithPreset(PresetReadOnly, "viewer123", "org123")
```

### Permission Matrix

Configure custom permissions:

```go
mock := factory.NewAAAServiceMock()
matrix := mock.GetPermissionMatrix()

// Add allow rules
matrix.AddAllowRule("user123", "farmer", "create", "*", "org123")
matrix.AddAllowRule("user123", "farm", "read", "*", "org123")

// Add deny rules (first match wins)
matrix.AddDenyRule("user123", "farmer", "delete", "*", "org123")

// Wildcard matching
matrix.AddAllowRule("admin", "*", "*", "*", "org123") // Admin can do anything

// Clear all rules
matrix.Clear()
```

### Permission Modes

```go
// Deny by default (most secure)
secureFactory := NewMockFactory(PermissionModeDenyAll)

// Allow all (for simple tests only)
permissiveFactory := NewMockFactory(PermissionModeAllowAll)

// Custom mode (requires explicit rules)
customFactory := NewMockFactory(PermissionModeCustom)
```

## Best Practices

### 1. Test Isolation

Each test should be independent:

```go
func TestExample(t *testing.T) {
    // Create fresh mock for each test
    mock := NewMockAAAServiceShared(true)

    // Test...

    // No cleanup needed - each test gets fresh state
}
```

### 2. Use Short Mode for CI

Mark integration tests to skip in short mode:

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    // ...
}
```

Run in CI:
```bash
go test ./... -short
```

### 3. Parallel Test Execution

Use parallel tests for better performance:

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // Test implementation
}
```

### 4. Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestPermissions(t *testing.T) {
    tests := []struct {
        name     string
        user     string
        resource string
        expected bool
    }{
        {"admin can create", "admin", "farmer", true},
        {"farmer cannot delete", "farmer", "farmer", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 5. Security-First Testing

Always test with deny-by-default:

```go
// Good: Secure by default
mock := NewMockAAAServiceShared(true)

// Bad: Too permissive for security tests
mock := NewMockAAAServiceShared(false)
```

### 6. Verify Audit Logs

Always verify security-critical operations are logged:

```go
mock, _ := NewSecurityEnhancedMockAAA(true)

// Perform operation
mock.CheckPermission(...)

// Verify it was logged
events := mock.GetAuditEvents()
assert.Greater(t, len(events), 0)
```

### 7. Test Error Paths

Don't just test happy paths:

```go
// Test error conditions
_, err := mock.CheckPermission(ctx, "", "", "", "", "")
assert.Error(t, err)

// Test edge cases
_, err = mock.CheckPermission(ctx, "user", "resource", "action", "", "org")
// Should handle empty object gracefully
```

### 8. Use Realistic Test Data

Use realistic IDs and data:

```go
// Good
userID := "usr_2Nk8HRk3X9vU"
orgID := "org_1Hd93Kl8M2pL"

// Avoid
userID := "test"
orgID := "123"
```

### 9. Clean Up Resources

TestContainers handles cleanup automatically, but for other resources:

```go
func TestWithCleanup(t *testing.T) {
    resource := createResource()

    t.Cleanup(func() {
        resource.Close()
    })

    // Use resource...
}
```

### 10. Document Complex Tests

Add comments for complex test scenarios:

```go
func TestComplexScenario(t *testing.T) {
    // This test validates that when a farmer is reassigned to a different FPO,
    // their existing farm permissions are revoked but farm ownership is preserved.
    // This prevents permission leakage across organizations.

    // Setup...
}
```

## Troubleshooting

### TestContainers Issues

**Problem**: Container fails to start

```bash
# Check Docker is running
docker ps

# Check Docker logs
docker logs <container-id>

# Increase timeout
export TESTCONTAINERS_RYUK_DISABLED=true
```

**Problem**: Tests are slow

```bash
# Use parallel tests
go test ./... -parallel 4

# Use test-short for fast feedback
make test-short
```

### Permission Test Failures

**Problem**: Permissions not working as expected

```go
// Enable debug logging
mock.GetPermissionMatrix().logDenials = true

// Check rule order (first match wins)
matrix.Clear()
matrix.AddAllowRule(...) // Add in correct order
```

### Rate Limit Issues

**Problem**: Rate limits triggering in tests

```go
// Disable rate limiting for specific tests
mock.GetRateLimiter().Disable()

// Or reset between tests
mock.GetRateLimiter().Reset("key")
```

## Contributing

When adding new tests:

1. Follow existing patterns and conventions
2. Add contract tests for new mock implementations
3. Include security tests for new security features
4. Update this documentation
5. Ensure all tests pass before committing

## Resources

- [TestContainers Documentation](https://testcontainers.com/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [PostGIS Documentation](https://postgis.net/docs/)
- [JWT Documentation](https://jwt.io/)
