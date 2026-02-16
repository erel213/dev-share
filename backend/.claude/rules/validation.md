---
paths:
  - "pkg/validation/**/*.go"
  - "pkg/contracts/**/*.go"
  - "internal/application/**/*.go"
---

# Request Validation

## Quick Reference

**Validation Flow:** Handler → Service validates → Domain → Repository

**Where:** Application layer (`internal/application/`) - first line of every service method that accepts a contract.

**Error Response:** HTTP 400 with field-level errors in metadata.

## Core Pattern

### In Services (REQUIRED)

```go
func (s UserService) CreateLocalUser(ctx context.Context, request contracts.CreateLocalUser) *errors.Error {
    // ✅ ALWAYS validate first
    if err := s.validator.Validate(request); err != nil {
        return err
    }
    // ... business logic
}
```

### Service Setup

```go
type UserService struct {
    validator *validation.Service
    // ... other fields
}

func NewUserService(validator *validation.Service, ...) UserService {
    return UserService{validator: validator, ...}
}
```

### Initialization (main.go)

```go
validator := validation.New()
if err := validator.RegisterDefaultCustomValidations(); err != nil {
    log.Fatal(err)
}
```

## Validation Tags

**Common patterns:**

```go
type CreateUser struct {
    Name        string    `json:"name" validate:"required,min=2,max=100"`
    Email       string    `json:"email" validate:"required,email"`
    Password    string    `json:"password" validate:"required,min=8,strongpassword"`
    WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
    Age         int       `json:"age" validate:"gte=18,lte=120"`
    Role        string    `json:"role" validate:"oneof=admin user guest"`
    Bio         *string   `json:"bio" validate:"omitempty,max=500"`  // Optional
}
```

**Tag quick reference:**
- `required` - Field must be present
- `min=N`, `max=N` - String length or numeric value
- `email` - Valid email format
- `uuid4` - Valid UUID v4
- `oneof=a b c` - Enum validation
- `gte`, `lte`, `gt`, `lt` - Numeric comparisons
- `omitempty` - Optional field (use with pointer types)
- `dive` - Validate nested structs/slices

**Custom validators:**
- `strongpassword` - Uppercase, lowercase, digit, special char (combine with `min=8`)

## Critical Rules

### ✅ DO

1. **Validate in service layer** - First line of service methods
2. **Use JSON tag names** - `json:"email"` tag is required
3. **Return all errors** - Validator returns all field errors at once
4. **Use pointers for optional** - `*string` with `omitempty` for optional fields

### ❌ DON'T

1. **Don't validate in handlers** - Service layer only
2. **Don't validate in domain** - Domain assumes valid data
3. **Don't manually check rules** - Use validator, not `if len(password) < 8`
4. **Don't skip validation** - Even for "trusted" inputs

## Custom Validators

To add a custom validator:

1. Add function to `pkg/validation/custom.go`:
```go
func validateCustom(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    return isValid(value)
}
```

2. Register in `RegisterDefaultCustomValidations()`:
```go
if err := s.RegisterCustomValidation("customtag", validateCustom); err != nil {
    return err
}
```

3. Add error message to `formatValidationError()` in `validator.go`:
```go
case "customtag":
    return field + " must meet custom requirements"
```

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION",
    "message": "validation failed",
    "metadata": {
      "fields": {
        "email": "email must be a valid email address",
        "password": "password must be at least 8 characters"
      }
    }
  }
}
```

## Testing

```go
func TestService_ValidRequest(t *testing.T) {
    validator := validation.New()
    validator.RegisterDefaultCustomValidations()
    service := NewService(validator)

    err := service.Method(ctx, validRequest)
    assert.NoError(t, err)
}

func TestService_InvalidRequest(t *testing.T) {
    err := service.Method(ctx, invalidRequest)
    assert.Equal(t, errors.CodeValidation, err.Code())

    fields := err.Metadata()["fields"].(map[string]string)
    assert.Contains(t, fields, "email")
}
```

## Complete Example

```go
// Contract
type CreateWorkspace struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Plan  string `json:"plan" validate:"required,oneof=free pro enterprise"`
}

// Service
func (s WorkspaceService) Create(ctx context.Context, req contracts.CreateWorkspace) *errors.Error {
    if err := s.validator.Validate(req); err != nil {
        return err
    }
    // Business logic...
}

// Handler
func (h Handler) Create(c *fiber.Ctx) error {
    var req contracts.CreateWorkspace
    if err := c.BodyParser(&req); err != nil {
        return err
    }
    return h.service.Create(c.Context(), req)  // Service validates
}
```

## Key Takeaway

**Validation = Service Layer Responsibility**

One call to `s.validator.Validate(request)` at the start of every service method. That's it.
