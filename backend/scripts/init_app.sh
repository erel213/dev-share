#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="dev-share"
POSTGRES_CONTAINER="dev-share-postgres"
BACKEND_CONTAINER="dev-share-backend"

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to clean volumes
clean_volumes() {
    print_info "Stopping and removing containers and volumes..."
    docker-compose -f "$COMPOSE_FILE" down -v
    print_info "Volumes cleaned successfully"
}

# Function to wait for postgres to be healthy
wait_for_postgres() {
    print_info "Waiting for PostgreSQL to be healthy..."
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker exec "$POSTGRES_CONTAINER" pg_isready -U devshare > /dev/null 2>&1; then
            print_info "PostgreSQL is ready!"
            return 0
        fi

        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done

    print_error "PostgreSQL failed to become healthy after $max_attempts seconds"
    return 1
}

# Function to run database migrations
run_migrations() {
    print_info "Running database migrations..."

    # Check if migrations directory exists
    if [ ! -d "internal/infra/migrations" ]; then
        print_warn "Migrations directory not found at internal/infra/migrations"
        return 0
    fi

    # Check if there are any migration files
    if [ -z "$(ls -A internal/infra/migrations 2>/dev/null)" ]; then
        print_warn "No migration files found in internal/infra/migrations"
        return 0
    fi

    # Run migrations using golang-migrate from within the backend container
    # This assumes the container has the migrate tool or we run it via the app
    docker exec "$POSTGRES_CONTAINER" psql -U devshare -d devshare -c "SELECT version();" > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        print_info "Database connection verified"

        # Option 1: If you have migrate CLI installed locally
        if command -v migrate &> /dev/null; then
            migrate -path internal/infra/migrations \
                    -database "postgres://devshare:devshare_password@localhost:5432/devshare?sslmode=disable" \
                    up
            print_info "Migrations completed successfully"
        else
            print_warn "golang-migrate CLI not found. Please install it or run migrations manually:"
            print_warn "  migrate -path internal/infra/migrations -database \"postgres://devshare:devshare_password@localhost:5432/devshare?sslmode=disable\" up"
        fi
    else
        print_error "Failed to connect to database"
        return 1
    fi
}

# Function to build and start services
start_services() {
    print_info "Building backend service image..."
    docker-compose -f "$COMPOSE_FILE" build backend

    print_info "Starting services..."
    docker-compose -f "$COMPOSE_FILE" up -d

    wait_for_postgres

    print_info "Waiting for backend service to start..."
    sleep 5

    # Check backend health
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_info "Backend service is healthy!"
            return 0
        fi

        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done

    print_warn "Backend service may not be fully ready. Check logs with: docker-compose logs backend"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    --clean-volumes    Remove all volumes before starting (WARNING: This will delete all data!)
    -h, --help         Show this help message

Description:
    This script initializes the Dev-Share backend application by:
    1. Building the backend service Docker image
    2. Starting all services via docker-compose
    3. Running database migrations

Example:
    $0                    # Start normally
    $0 --clean-volumes    # Clean volumes and start fresh
EOF
}

# Main execution
main() {
    local clean_volumes_flag=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --clean-volumes)
                clean_volumes_flag=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    print_info "Initializing Dev-Share Backend Application..."

    # Clean volumes if requested
    if [ "$clean_volumes_flag" = true ]; then
        print_warn "WARNING: This will remove all data!"
        read -p "Are you sure you want to continue? (yes/no): " confirmation
        if [ "$confirmation" = "yes" ]; then
            clean_volumes
        else
            print_info "Aborted by user"
            exit 0
        fi
    fi

    # Start services
    start_services

    # Run migrations
    run_migrations

    print_info "========================================="
    print_info "Application initialized successfully!"
    print_info "========================================="
    print_info "Backend API: http://localhost:8080"
    print_info "PostgreSQL: localhost:5432"
    print_info ""
    print_info "Useful commands:"
    print_info "  docker-compose logs -f backend    # View backend logs"
    print_info "  docker-compose logs -f postgres   # View postgres logs"
    print_info "  docker-compose ps                 # View running services"
    print_info "  docker-compose down               # Stop services"
    print_info "========================================="
}

# Run main function
main "$@"
