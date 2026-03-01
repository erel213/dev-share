#!/bin/bash
# Remove all SQLite database files
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Removing SQLite database files..."
find "$SCRIPT_DIR" -type f \( -name "*.db" -o -name "*.db-wal" -o -name "*.db-shm" \) -print -delete
echo "Done."
