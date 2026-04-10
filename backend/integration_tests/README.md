# Integration Tests

This directory contains integration tests for the dev-share backend. The tests interact with real HTTP endpoints against a backend running with SQLite.

## Overview

The integration test suite includes:

- **Admin Init Tests** (`admin_init_test.go`) — First-time system initialization, token protection
- **Admin User Management Tests** (`admin_user_management_test.go`) — User invitations, password resets, user deletion
- **Group Tests** (`group_test.go`) — Group CRUD, member management, template access control
- **Login Tests** (`login_test.go`) — Authentication, JWT cookie handling
- **Template Tests** (`template_test.go`) — Template CRUD, file upload, variable parsing
- **User Tests** (`user_test.go`) — User creation, validation, cascade deletes
- **Workspace Tests** (`workspace_test.go`) — CRUD operations, pagination, validation

## Quick Start

### Using Make (Recommended)

```bash
# Run all integration tests
make test-integration

# Run specific test suites
make test-workspace
make test-user
make test-admin
make test-template
make test-group

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
./scripts/run-integration-tests.sh admin
./scripts/run-integration-tests.sh template
./scripts/run-integration-tests.sh group

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
# Run tests directly
cd backend
go test ./integration_tests/... -v -timeout 120s

# Run a specific test file
go test ./integration_tests/... -run TestCreateWorkspace_Success -v
```

## Test Environment

The tests run the backend with SQLite and connect via HTTP.

### Environment Variables

You can customize the test environment:

```bash
# Change the base URL for tests
export TEST_BASE_URL="http://localhost:8081"

# Then run tests
./scripts/run-integration-tests.sh
```

## Test Structure

### helpers_test.go

Provides HTTP client wrappers for all API endpoints:

**Workspace Helpers:**
- `CreateWorkspace(t, auth, name, description, adminID)`
- `GetWorkspace(t, auth, id)`
- `GetWorkspacesByAdmin(t, auth, adminID)`
- `UpdateWorkspace(t, auth, id, name, description)`
- `DeleteWorkspace(t, auth, id)`
- `ListWorkspaces(t, auth, limit, offset, sortBy, order)`

**User Helpers:**
- `CreateUser(t, name, email, password, workspaceID)`
- `LoginUser(t, email, password)`

**Admin Helpers:**
- `InitializeAdmin(t, adminName, adminEmail, adminPassword, workspaceName, workspaceDescription, token)`
- `AdminInviteUser(t, auth, name, email, role)`
- `AdminResetUserPassword(t, auth, userID)`
- `AdminListUsers(t, auth)`
- `AdminDeleteUser(t, auth, userID)`

**Template Helpers:**
- `CreateTemplate(t, auth, name, workspaceID, files)`
- `CreateTemplateRaw(t, auth, name, workspaceID, files)`
- `GetTemplate(t, auth, id)`
- `GetTemplatesByWorkspace(t, auth, workspaceID)`
- `UpdateTemplate(t, auth, id, name, files...)`
- `DeleteTemplate(t, auth, id)`
- `ListTemplateFiles(t, auth, templateID)`
- `GetTemplateFileContent(t, auth, templateID, path)`
- `ListTemplates(t, auth, limit, offset, sortBy, order)`

**Group Helpers:**
- `CreateGroup(t, auth, name, description, accessAllTemplates)`
- `GetGroup(t, auth, id)`
- `ListGroups(t, auth)`
- `UpdateGroup(t, auth, id, payload)`
- `DeleteGroup(t, auth, id)`
- `AddGroupMembers(t, auth, groupID, userIDs)`
- `GetGroupMembers(t, auth, groupID)`
- `RemoveGroupMember(t, auth, groupID, userID)`
- `AddGroupTemplateAccess(t, auth, groupID, templateIDs)`
- `GetGroupTemplateAccess(t, auth, groupID)`
- `RemoveGroupTemplateAccess(t, auth, groupID, templateID)`

All helpers accept an `AuthContext` for JWT-authenticated requests and return the response struct plus HTTP status code for assertion.

## Writing New Tests

### Example Test

```go
func TestMyFeature_Success(t *testing.T) {
    // Setup: Create prerequisites
    auth := AuthContext{
        UserID:      adminUserID,
        UserName:    "Admin",
        Role:        "admin",
        WorkspaceID: workspaceID,
    }
    workspace, _ := CreateWorkspace(t, auth, "Test Workspace", "Description", auth.UserID)

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

Ensure the test port is not in use:
```bash
lsof -i :8081
```

### Migration failures

Check migrations are valid:
```bash
ls -la internal/infra/migrations/sqlite/
```

Ensure migrations are numbered sequentially and have both `.up.sql` and `.down.sql` files.
