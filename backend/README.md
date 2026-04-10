# Dev-Share Backend

Backend service for managing developer environments, enabling clients to provision, configure, and control their development infrastructure through IaC (Terraform) integration.

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) - Express-inspired HTTP framework
- **Database**: SQLite (file-based, via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite))
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Architecture**: Clean architecture with domain-driven design
- **IaC Engine**: Terraform (executed server-side for environment provisioning)

## Getting Started

### Prerequisites
- Go 1.24+

### Running Locally

The recommended way is to use the root-level setup script:

```bash
# From the repo root
./setup.sh
```

Or run manually:

```bash
cd backend
make run
```

See `make help` for all available Make targets.

### Database Migrations

Migrations are located in `internal/infra/migrations/sqlite/`. Each migration consists of two files:
- `{version}_{description}.up.sql` — applied when migrating up
- `{version}_{description}.down.sql` — applied when rolling back

Migrations run automatically on startup. To create a new migration:

```bash
make db-migrate-create name=<migration_name>
```

### Make Targets

| Target | Description |
|---|---|
| `make run` | Run the backend locally |
| `make build` | Build the backend binary to `bin/server` |
| `make lint` | Run `go vet` and `go fmt` |
| `make deps` | Download and tidy dependencies |
| `make test-unit` | Run unit tests |
| `make test-integration` | Run all integration tests |
| `make test-workspace` | Run workspace integration tests |
| `make test-user` | Run user integration tests |
| `make test-admin` | Run admin integration tests |
| `make test-template` | Run template integration tests |
| `make test-group` | Run group integration tests |
| `make test-all` | Run all tests (unit + integration) |
| `make db-migrate-create name=...` | Create a new migration |

## API Endpoints

All resource endpoints are prefixed with `/api/v1`.

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/admin/status` | System initialization status |
| `POST` | `/admin/init` | First-time system setup (admin + workspace) |
| `GET` | `/api/v1/` | API version info |
| `POST` | `/api/v1/users` | Register a new user |
| `POST` | `/api/v1/login` | Log in (sets httpOnly JWT cookie) |

### Authenticated

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/me` | Current user info |

### Environments (all authenticated users)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/environments` | Create environment |
| `GET` | `/api/v1/environments` | List environments |
| `GET` | `/api/v1/environments/:id` | Get environment |
| `POST` | `/api/v1/environments/:id/plan` | Run Terraform plan |
| `POST` | `/api/v1/environments/:id/apply` | Run Terraform apply |
| `POST` | `/api/v1/environments/:id/destroy` | Destroy environment resources |
| `DELETE` | `/api/v1/environments/:id` | Delete environment |
| `PUT` | `/api/v1/environments/:id/variables` | Set variable values |
| `GET` | `/api/v1/environments/:id/variables` | Get variable values |

### Workspaces (editor+ can write, all can read)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/workspaces` | Create workspace |
| `GET` | `/api/v1/workspaces` | List workspaces |
| `GET` | `/api/v1/workspaces/:id` | Get workspace |
| `GET` | `/api/v1/workspaces/admin/:admin_id` | Get workspaces by admin |
| `PUT` | `/api/v1/workspaces/:id` | Update workspace |
| `DELETE` | `/api/v1/workspaces/:id` | Delete workspace |

### Templates (editor+ can write, all can read)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/templates` | Create template (multipart upload) |
| `GET` | `/api/v1/templates` | List templates |
| `GET` | `/api/v1/templates/:id` | Get template |
| `GET` | `/api/v1/templates/workspace/:workspace_id` | Get templates by workspace |
| `PUT` | `/api/v1/templates/:id` | Update template |
| `DELETE` | `/api/v1/templates/:id` | Delete template |
| `GET` | `/api/v1/templates/:id/files` | List template files |
| `GET` | `/api/v1/templates/:id/files/content` | Get template file content |

### Template Variables (editor+ can write, all can read)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/templates/:id/variables` | Create variable |
| `GET` | `/api/v1/templates/:id/variables` | List variables |
| `PUT` | `/api/v1/templates/:id/variables/:varId` | Update variable |
| `DELETE` | `/api/v1/templates/:id/variables/:varId` | Delete variable |
| `POST` | `/api/v1/templates/:id/variables/parse` | Parse and reconcile variables from template files |

### Groups (admin only)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/groups` | Create group |
| `GET` | `/api/v1/groups` | List groups |
| `GET` | `/api/v1/groups/:id` | Get group |
| `PUT` | `/api/v1/groups/:id` | Update group |
| `DELETE` | `/api/v1/groups/:id` | Delete group |
| `POST` | `/api/v1/groups/:id/members` | Add members |
| `GET` | `/api/v1/groups/:id/members` | Get members |
| `DELETE` | `/api/v1/groups/:id/members/:user_id` | Remove member |
| `POST` | `/api/v1/groups/:id/templates` | Add template access |
| `GET` | `/api/v1/groups/:id/templates` | Get template access |
| `DELETE` | `/api/v1/groups/:id/templates/:template_id` | Remove template access |

### Admin User Management (admin only)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/admin/users` | List all users |
| `POST` | `/api/v1/admin/users/invite` | Invite a new user |
| `POST` | `/api/v1/admin/users/:id/reset-password` | Reset user password |
| `DELETE` | `/api/v1/admin/users/:id` | Delete user |

## Error Handling

The error handling system is organized into layers:

1. **Foundation Layer** (`pkg/errors`)
   - Core `Error` type with metadata, severity levels, and error codes
   - Smart stack trace capture (only for Error/Critical severity)
   - Full `log/slog` integration for structured logging

2. **Domain Layer** (`internal/domain/errors`)
   - Domain-specific error constructors with entity context
   - `NotFound(entityType, id)` — entity not found errors
   - `Conflict(entityType, field, value)` — unique constraint violations
   - `InvalidInput(field, reason)` — validation errors

3. **Infrastructure Layer** (`internal/infra/errors`)
   - Database error mapping (`WrapDatabaseError`)
   - Automatic SQLite constraint violation detection

4. **HTTP Layer** (`internal/application/errors`)
   - Fiber error middleware for automatic error-to-HTTP conversion
   - JSON error responses with proper status codes
   - Structured error logging with request context

### Error Response Format

All HTTP errors return JSON in this format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found: 123e4567-e89b-12d3-a456-426614174000",
    "metadata": {
      "entity_type": "User",
      "entity_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

For detailed guidelines on using the error handling system, see `.claude/rules/error-handling.md`.

## Architecture

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Domain models, repository interfaces, errors
│   ├── application/     # HTTP handlers, service orchestration, error middleware
│   └── infra/           # SQLite, Terraform execution, file storage, migrations
└── pkg/                 # Shared packages (config, JWT, crypto, validation, errors)
```

Dependencies point inward: `infra` → `application` → `domain`. The `pkg/` layer is shared across all layers.
