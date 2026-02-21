#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/internal/infra/migrations"

if [ -z "$1" ]; then
    echo "Usage: $(basename "$0") <migration_name>"
    echo "Example: $(basename "$0") add_email_to_users"
    exit 1
fi

MIGRATION_NAME="$1"

if ! command -v migrate &> /dev/null; then
    echo "Error: golang-migrate CLI is not installed"
    echo "Install with: brew install golang-migrate"
    exit 1
fi

migrate create -ext sql -dir "$MIGRATIONS_DIR" -seq "$MIGRATION_NAME"

echo "Created migration: $MIGRATION_NAME"
