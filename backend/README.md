# Dev-Share Backend

Backend service for managing developer environments, enabling clients to provision, configure, and control their development infrastructure.

## Tech Stack

- **Language**: Go 1.23.4
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) - Express-inspired HTTP framework
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate) - Schema migration management
- **Architecture**: Clean architecture with domain-driven design
- **Error Handling**: Robust layered error system with structured logging and observability

## Getting Started

### Prerequisites
- Go 1.23.4 or higher
- Docker (optional)

### Running Locally
#### Quick Start with Init Script (Recommended)

The easiest way to run the application is using the initialization script, which handles building, starting services, and running migrations:

```bash
# Make the script executable (first time only)
chmod +x scripts/init_app.sh

# Run the application
./scripts/init_app.sh

# Start fresh with clean volumes (removes all data)
./scripts/init_app.sh --clean-volumes
```

The script will:
1. Build the backend service Docker image
2. Start all services (backend + PostgreSQL) via docker-compose
3. Wait for PostgreSQL to be healthy
4. Run database migrations automatically
5. Verify the backend service is running

### Database Migrations

We use [golang-migrate](https://github.com/golang-migrate/migrate) for managing database schema changes.

#### Install Migration CLI

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows
scoop install migrate
```

#### Migration Commands

```bash
# Run all pending migrations
migrate -path internal/infra/migrations -database "postgresql://user:password@localhost:5432/devshare?sslmode=disable" up

# Rollback the last migration
migrate -path internal/infra/migrations -database "postgresql://user:password@localhost:5432/devshare?sslmode=disable" down 1

# Create a new migration
migrate create -ext sql -dir internal/infra/migrations -seq <migration_name>

# Check migration version
migrate -path internal/infra/migrations -database "postgresql://user:password@localhost:5432/devshare?sslmode=disable" version
```

#### Migration Files Location

All migration files are located in `internal/infra/migrations/`. Each migration consists of two files:
- `{version}_{description}.up.sql` - Applied when migrating up
- `{version}_{description}.down.sql` - Applied when rolling back

### Error Handling

The error handling system is organized into layers:

1. **Foundation Layer** (`pkg/errors`)
   - Core `Error` type with metadata, severity levels, and error codes
   - Smart stack trace capture (only for Error/Critical severity)
   - Full `log/slog` integration for structured logging

2. **Domain Layer** (`internal/domain/errors`)
   - Domain-specific error constructors with entity context
   - `NotFound(entityType, id)` - Entity not found errors
   - `Conflict(entityType, field, value)` - Unique constraint violations
   - `InvalidInput(field, reason)` - Validation errors

3. **Infrastructure Layer** (`internal/infra/errors`)
   - Database error mapping (`WrapDatabaseError`)
   - Automatic PostgreSQL constraint violation detection
   - Transaction and connection error utilities

4. **HTTP Layer** (`internal/handler/errors`)
   - Fiber error middleware for automatic error-to-HTTP conversion
   - JSON error responses with proper status codes
   - Structured error logging with request context

#### Key Features

- **Rich Context**: Every error includes metadata, timestamps, and severity levels
- **Smart Stack Traces**: Captured only for unexpected errors (Error/Critical severity)
- **Structured Logging**: Full integration with `log/slog` for JSON logging
- **HTTP Integration**: Automatic error-to-HTTP response conversion with proper status codes
- **PostgreSQL Intelligence**: Automatic constraint violation detection and classification
- **Backward Compatible**: Existing code using old error types continues to work

#### Error Response Format

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

### API Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/` - API version information
