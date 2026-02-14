# Dev-Share Backend

Backend service for managing developer environments, enabling clients to provision, configure, and control their development infrastructure.

## Tech Stack

- **Language**: Go 1.23.4
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) - Express-inspired HTTP framework
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate) - Schema migration management
- **Architecture**: Clean architecture with domain-driven design

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

### API Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/` - API version information
