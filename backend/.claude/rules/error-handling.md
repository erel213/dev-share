---
paths:
  - "pkg/errors/**/*.go"
  - "internal/domain/errors/**/*.go"
  - "internal/infra/errors/**/*.go"
  - "internal/handler/errors/**/*.go"
  - "internal/infra/postgres/**/*.go"
  - "internal/handler/**/*.go"
---

# Error Handling Rules

## Overview

This codebase uses a robust, layered error handling system with rich context and observability. Errors flow through multiple layers, each adding appropriate context and handling.

## Architecture Layers

### 1. Foundation Layer (`pkg/errors`)

**Core error type with:**
- Error codes (NOT_FOUND, CONFLICT, INVALID_INPUT, etc.)
- Severity levels (Debug, Info, Warning, Error, Critical)
- Metadata for context
- Stack traces (captured only for Error/Critical severity)
- HTTP status code mapping
- `log/slog` integration

**DO NOT** modify this layer without careful consideration - it's the foundation for all error handling.

### 2. Domain Layer (`internal/domain/errors`)

**Use domain error constructors for all domain-level errors:**

```go
// Entity not found (severity: Warning)
domainerrors.NotFound(entityType string, id uuid.UUID) *pkgerrors.Error

// Lookup by field instead of UUID
domainerrors.NotFoundByField(entityType, field, value string) *pkgerrors.Error

// Unique constraint violations (severity: Warning)
domainerrors.Conflict(entityType, field, value string) *pkgerrors.Error

// Validation errors (severity: Warning)
domainerrors.InvalidInput(field, reason string) *pkgerrors.Error

// Multiple field validation errors
domainerrors.ValidationError(message string, fieldErrors map[string]string) *pkgerrors.Error

// Authentication errors (severity: Warning)
domainerrors.Unauthorized(reason string) *pkgerrors.Error

// Authorization errors (severity: Warning)
domainerrors.Forbidden(resource, action string) *pkgerrors.Error
```

**When to use:**
- Business logic violations
- Entity not found scenarios
- Validation failures
- Domain-specific error conditions

### 3. Infrastructure Layer (`internal/infra/errors`)

**Database Error Handling:**

```go
// Wrap ALL database errors - automatically classifies and adds context
infraerrors.WrapDatabaseError(err error, operation string) error

// Use for transaction errors
infraerrors.WrapTransactionError(err error, operation string) error

// Use for connection errors (severity: Critical)
infraerrors.WrapConnectionError(err error) error
```

**CRITICAL Rules for Repository Implementations:**

1. **ALWAYS wrap database errors** - Even if you think you know what the error is:
   ```go
   // ❌ WRONG - manual error handling
   if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
       return &repository.ConflictError{...}
   }

   // ✅ CORRECT - use WrapDatabaseError
   return infraerrors.WrapDatabaseError(err, "create_user")
   ```

2. **Use descriptive operation names** - They appear in logs and help debugging:
   ```go
   infraerrors.WrapDatabaseError(err, "get_user_by_email")
   infraerrors.WrapDatabaseError(err, "update_workspace")
   infraerrors.WrapDatabaseError(err, "list_environments")
   ```

3. **Handle sql.ErrNoRows explicitly** (when you need entity context):
   ```go
   err = r.db.QueryRowContext(ctx, query, args...).Scan(...)
   if err != nil {
       if err == sql.ErrNoRows {
           return domainerrors.NotFound("User", id)  // Rich context
       }
       return infraerrors.WrapDatabaseError(err, "get_user")
   }
   ```

4. **For queries that may or may not find results**, use WrapDatabaseError directly:
   ```go
   // WrapDatabaseError automatically handles sql.ErrNoRows as NOT_FOUND
   if err := r.db.QueryRowContext(...).Scan(...); err != nil {
       return infraerrors.WrapDatabaseError(err, "get_user_by_oauth")
   }
   ```

### 4. HTTP Layer (`internal/handler/errors`)

**Error Middleware:**

The error handler middleware is automatically registered in `fiber.Config.ErrorHandler`. It:
- Converts errors to JSON responses
- Logs errors with structured context
- Maps error codes to HTTP status codes

**Handler Error Returns:**

```go
// ✅ Simply return errors - middleware handles conversion
user, err := userRepo.GetByID(ctx, id)
if err != nil {
    return err  // Automatically becomes proper HTTP response
}

// ✅ Use helper functions for direct errors
if invalidID {
    return handlererrors.ReturnBadRequest("invalid user ID format")
}

if notFound {
    return handlererrors.ReturnNotFound("user not found")
}

// Available helpers:
handlererrors.ReturnNotFound(message string)
handlererrors.ReturnBadRequest(message string)
handlererrors.ReturnConflict(message string)
handlererrors.ReturnUnauthorized(message string)
handlererrors.ReturnForbidden(message string)
handlererrors.ReturnInternalError(message string)
```

## Severity Guidelines

Choose appropriate severity levels:

- **Debug**: Development/troubleshooting info (no stack trace)
- **Info**: Normal operation events (no stack trace)
- **Warning**: Expected error conditions - validation failures, not found, conflicts (no stack trace)
- **Error**: Unexpected errors requiring investigation (stack trace captured)
- **Critical**: System failures, connection errors (stack trace captured)

**Default severities:**
- `NotFound`, `Conflict`, `InvalidInput` → Warning (expected conditions)
- `Unauthorized`, `Forbidden` → Warning (expected auth failures)
- Unknown database errors → Error (needs investigation)
- Connection failures → Critical (system failure)

## DO NOT

1. **DO NOT** create raw errors with `errors.New()` or `fmt.Errorf()` in domain/handler layers
   ```go
   // ❌ WRONG
   return fmt.Errorf("user not found: %s", id)

   // ✅ CORRECT
   return domainerrors.NotFound("User", id)
   ```

2. **DO NOT** manually check PostgreSQL error codes
   ```go
   // ❌ WRONG - WrapDatabaseError handles this
   if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
       // manual handling
   }

   // ✅ CORRECT
   return infraerrors.WrapDatabaseError(err, "create_user")
   ```

3. **DO NOT** wrap errors multiple times
   ```go
   // ❌ WRONG - double wrapping
   err = infraerrors.WrapDatabaseError(err, "get_user")
   return pkgerrors.Wrap(err, "failed to get user")

   // ✅ CORRECT - single wrap
   return infraerrors.WrapDatabaseError(err, "get_user")
   ```

4. **DO NOT** ignore error context
   ```go
   // ❌ WRONG - losing context
   if err != nil {
       return domainerrors.NotFound("User", uuid.Nil)
   }

   // ✅ CORRECT - preserve context
   return infraerrors.WrapDatabaseError(err, "get_user")
   ```

5. **DO NOT** log errors in repositories or domain layer
   ```go
   // ❌ WRONG - logging in repository
   if err != nil {
       log.Printf("failed to get user: %v", err)
       return err
   }

   // ✅ CORRECT - return error, let middleware log
   if err != nil {
       return infraerrors.WrapDatabaseError(err, "get_user")
   }
   ```

## Adding Metadata

Add contextual metadata to errors when helpful for debugging:

```go
err := domainerrors.NotFound("User", id)
return err.
    WithMetadata("request_ip", requestIP).
    WithMetadata("attempted_at", time.Now()).
    WithMetadata("correlation_id", correlationID)
```

**Common metadata to include:**
- Entity identifiers
- Field names and values
- Operation details
- Request/correlation IDs
- Timestamps

## Migration from Legacy Errors

**Deprecated types** (backward compatible, but prefer new system):
- `repository.NotFoundError` → Use `domainerrors.NotFound()`
- `repository.ConflictError` → Use `domainerrors.Conflict()`
- `repository.ErrNotFound` → Use `domainerrors.ErrNotFound`

**Checking error types:**
```go
// ✅ CORRECT - works with both old and new errors
if pkgerrors.IsNotFound(err) {
    // handle not found
}

// Also works:
if errors.Is(err, domainerrors.ErrNotFound) {
    // handle not found
}
```

## Testing Error Handling

When writing tests for error scenarios:

```go
// Test error codes
var appErr *pkgerrors.Error
if !errors.As(err, &appErr) {
    t.Fatal("expected application error")
}
assert.Equal(t, pkgerrors.CodeNotFound, appErr.Code())

// Test error detection
assert.True(t, pkgerrors.IsNotFound(err))

// Test metadata
metadata := appErr.GetMetadata()
assert.Equal(t, "User", metadata["entity_type"])
```

## Logging

All errors are automatically logged by the HTTP middleware with:
- Error message and code
- Severity level
- HTTP status code
- Request path and method
- Metadata
- Stack trace (for Error/Critical severity)

**DO NOT** add manual logging in:
- Repository layer
- Domain layer
- Handler layer (errors)

Let the middleware handle logging consistently.

## Example: Complete Repository Method

```go
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
    query, args, err := StatementBuilder.
        Insert("users").
        Columns("name", "email").
        Values(user.Name, user.Email).
        Suffix("RETURNING id, created_at, updated_at").
        ToSql()
    if err != nil {
        // Query building should never fail in production
        return infraerrors.WrapDatabaseError(err, "build_create_user_query")
    }

    err = r.db.QueryRowContext(ctx, query, args...).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        // WrapDatabaseError automatically:
        // - Detects unique violations → CodeConflict (HTTP 409)
        // - Detects foreign key violations → CodeInvalidInput (HTTP 400)
        // - Unknown errors → CodeDatabase (HTTP 500) with stack trace
        return infraerrors.WrapDatabaseError(err, "create_user")
    }

    return nil
}
```

## When Creating New Error Types

If you need to create new domain errors:

1. Add them to `internal/domain/errors/errors.go`
2. Use descriptive function names
3. Set appropriate severity (usually Warning for business logic)
4. Set appropriate HTTP status code
5. Add relevant metadata
6. Follow the existing pattern

Example:
```go
func RateLimitExceeded(userID uuid.UUID, limit int) *pkgerrors.Error {
    return pkgerrors.WithCode(
        pkgerrors.CodeRateLimitExceeded,  // Add to pkg/errors/codes.go
        "rate limit exceeded",
    ).
        WithMetadata("user_id", userID.String()).
        WithMetadata("limit", limit).
        WithHTTPStatus(http.StatusTooManyRequests).
        WithSeverity(pkgerrors.SeverityWarning)
}
```
