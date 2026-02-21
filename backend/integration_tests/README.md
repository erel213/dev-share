# Integration Tests

This directory contains end-to-end integration tests for the dev-share backend. The tests interact with real HTTP endpoints against a real PostgreSQL database running in Docker.

## Overview

The integration test suite includes:

- **Workspace Tests** (`workspace_test.go`) - CRUD operations, pagination, validation
- **User Tests** (`user_test.go`) - User creation, validation, cascade deletes

## Quick Start

### Using Make (Recommended)

```bash
# Run all integration tests
make test-integration

# Run specific test suites
make test-workspace
make test-user

# Run with verbose output
make test-integration-v
make test-workspace-v
make test-user-v

# Keep containers running after tests (for debugging)
make test-keep
```

### Using the Bash Script

```bash
# Run all tests
./scripts/run-integration-tests.sh

# Run specific test suite
./scripts/run-integration-tests.sh workspace
./scripts/run-integration-tests.sh user

# Run with verbose output
./scripts/run-integration-tests.sh -v all

# Keep containers running after tests
./scripts/run-integration-tests.sh -k all

# Skip rebuilding Docker images
./scripts/run-integration-tests.sh -s all

# Custom timeout
./scripts/run-integration-tests.sh -t 180s all
```

### Manual Execution

```bash
# Start the test environment
docker compose -f docker-compose.test.yml up -d --build

# Run tests
cd backend
go test ./integration_tests/... -v -timeout 120s

# Or run specific test file
go test ./integration_tests/workspace_test.go ./integration_tests/test_setup.go ./integration_tests/test_helpers.go -v

# Tear down
docker compose -f docker-compose.test.yml down -v
```

## Test Environment

The integration tests use a separate Docker Compose environment (`docker-compose.test.yml`) with:

- **PostgreSQL**: Port 5433 (to avoid conflicts with dev database on 5432)
- **Backend API**: Port 8081 (to avoid conflicts with dev server on 8080)
- **Test Database**: `devshare_test`

### Environment Variables

You can customize the test environment:

```bash
# Change the base URL for tests
export TEST_BASE_URL="http://localhost:8081"

# Change the database connection
export TEST_DB_DSN="postgres://devshare:devshare_password@localhost:5433/devshare_test?sslmode=disable"

# Then run tests
./scripts/run-integration-tests.sh
```

## Test Structure

### test_setup.go

Contains `TestMain` which:
1. Waits for the backend `/health` endpoint to respond
2. Runs all database migrations using golang-migrate
3. Executes the test suite
4. Rolls back migrations for cleanup

### test_helpers.go

Provides HTTP client wrappers for all API endpoints:

**Workspace Helpers:**
- `CreateWorkspace(t, name, description, adminID)`
- `GetWorkspace(t, id)`
- `GetWorkspacesByAdmin(t, adminID)`
- `UpdateWorkspace(t, id, name, description)`
- `DeleteWorkspace(t, id)`
- `ListWorkspaces(t, limit, offset, sortBy, order)`

**User Helpers:**
- `CreateUser(t, name, email, password, workspaceID)`

All helpers return the response struct and HTTP status code for assertion.

### Test Files

Each test file contains focused test cases:

**workspace_test.go** (15 tests):
- Create workspace (success & validation)
- Get workspace (success & not found)
- Get workspaces by admin
- Update workspace
- Delete workspace
- List workspaces with pagination

**user_test.go** (6 tests):
- Create user (success)
- Duplicate email handling
- Invalid workspace validation
- Password strength validation
- Field validation
- Cascade delete behavior

## Writing New Tests

### Example Test

```go
func TestMyFeature_Success(t *testing.T) {
    // Setup: Create prerequisites
    adminID := uuid.New()
    workspace, _ := CreateWorkspace(t, "Test Workspace", "Description", adminID)

    // Execute: Call the endpoint
    user, status := CreateUser(t, "John Doe", "john@example.com", "SecureP@ss1!", workspace.ID)

    // Assert: Verify results
    if status != http.StatusCreated {
        t.Fatalf("expected status 201, got %d", status)
    }

    if user.UserID == uuid.Nil {
        t.Error("expected non-nil user ID")
    }
}
```

### Best Practices

1. **Use helpers**: Always use test helpers instead of raw HTTP calls
2. **Clean test names**: Use descriptive names like `TestFeature_Scenario`
3. **Isolate tests**: Each test should be independent and create its own data
4. **Assert status codes**: Always check HTTP status codes first
5. **Use t.Helper()**: Mark helper functions with `t.Helper()` for better error messages
6. **Test edge cases**: Include validation errors, not found scenarios, conflicts

## Debugging

### View Logs

```bash
# While tests are running (or with -k flag)
docker compose -f docker-compose.test.yml logs -f backend-test
docker compose -f docker-compose.test.yml logs -f postgres-test
```

### Keep Containers Running

```bash
# Run tests and keep containers up for debugging
./scripts/run-integration-tests.sh -k all

# Then manually inspect
docker compose -f docker-compose.test.yml ps
docker compose -f docker-compose.test.yml exec backend-test /bin/sh
docker compose -f docker-compose.test.yml exec postgres-test psql -U devshare -d devshare_test

# Clean up when done
docker compose -f docker-compose.test.yml down -v
```

### Run Individual Tests

```bash
# Run a specific test function
go test ./integration_tests/... -run TestCreateWorkspace_Success -v

# Run tests matching a pattern
go test ./integration_tests/... -run TestCreate -v
```

## CI/CD Integration

The tests can be integrated into CI pipelines:

```yaml
# GitHub Actions example
- name: Run Integration Tests
  run: |
    cd backend
    ./scripts/run-integration-tests.sh all
```

## Troubleshooting

### Tests timing out

Increase the timeout:
```bash
./scripts/run-integration-tests.sh -t 300s all
```

### Port conflicts

Ensure ports 5433 and 8081 are not in use:
```bash
lsof -i :5433
lsof -i :8081
```

### Docker issues

```bash
# Clean up all containers
docker compose -f docker-compose.test.yml down -v

# Rebuild images
docker compose -f docker-compose.test.yml build --no-cache
```

### Migration failures

Check migrations are valid:
```bash
ls -la internal/infra/migrations/
```

Ensure migrations are numbered sequentially and have both `.up.sql` and `.down.sql` files.
