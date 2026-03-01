#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_step()  { echo -e "\n${BLUE}${BOLD}==>${NC}${BOLD} $1${NC}"; }
print_ok()    { echo -e "  ${GREEN}✓${NC} $1"; }
print_warn()  { echo -e "  ${YELLOW}!${NC} $1"; }
print_error() { echo -e "  ${RED}✗${NC} $1"; }

fail() { print_error "$1"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

BACKEND_PID=""
FRONTEND_PID=""
cleanup() {
  if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null
  fi
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null
  fi
}
trap cleanup EXIT

# ── 1. Check prerequisites ──────────────────────────────────────────

print_step "Checking prerequisites"

if ! command -v go &>/dev/null; then
  fail "Go is not installed. Please install Go 1.24+ from https://go.dev/dl/"
fi

GO_VERSION=$(go version | grep -oE '[0-9]+\.[0-9]+' | head -1)
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)
if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 24 ]; }; then
  fail "Go 1.24+ required (found $GO_VERSION)"
fi
print_ok "Go $GO_VERSION"

if ! command -v pnpm &>/dev/null; then
  fail "pnpm is not installed. Install it: npm install -g pnpm"
fi
print_ok "pnpm $(pnpm --version)"

if ! command -v curl &>/dev/null; then
  fail "curl is not installed"
fi
print_ok "curl"

# ── 2. Install backend dependencies ─────────────────────────────────

print_step "Installing backend dependencies"
(cd backend && go mod download) || fail "Failed to download Go modules"
print_ok "Go modules downloaded"

# ── 3. Install frontend dependencies ────────────────────────────────

print_step "Installing frontend dependencies"
(cd frontend && pnpm install) || fail "Failed to install frontend dependencies"
print_ok "Frontend dependencies installed"

# ── 4. Build frontend ───────────────────────────────────────────────

print_step "Building frontend"
(cd frontend && pnpm build) || fail "Frontend build failed"
print_ok "Frontend built"

# ── 5. Set up environment ───────────────────────────────────────────

print_step "Setting up environment"

if [ ! -f .env ]; then
  cp .env.example .env
  JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
  else
    sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
  fi
  print_ok ".env created with generated JWT_SECRET"
else
  print_warn ".env already exists, skipping"
fi

# Export env vars for the backend
set -a
source .env
set +a

# DB_FILE_PATH in .env is relative to the repo root (e.g. ./backend/devshare.db),
# but Go commands run from backend/, so strip the leading ./backend/ prefix.
if [[ "$DB_FILE_PATH" == ./backend/* ]]; then
  export DB_FILE_PATH="./${DB_FILE_PATH#./backend/}"
fi

# ── 6. Run database migrations ──────────────────────────────────────

print_step "Running database migrations"
(cd backend && go run ./cmd/migrate) || fail "Database migration failed"
print_ok "Migrations applied"

# ── 7. Start backend ────────────────────────────────────────────────

print_step "Starting backend server"
(cd backend && go run ./cmd/server) &
BACKEND_PID=$!
print_ok "Backend starting (PID: $BACKEND_PID)"

# ── 8. Wait for healthy ─────────────────────────────────────────────

print_step "Waiting for backend to be ready"

PORT="${PORT:-8080}"
MAX_ATTEMPTS=30
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  if curl -s "http://localhost:${PORT}/health" >/dev/null 2>&1; then
    print_ok "Backend is healthy at http://localhost:${PORT}"
    break
  fi
  if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    fail "Backend process exited unexpectedly"
  fi
  sleep 1
  ATTEMPT=$((ATTEMPT + 1))
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
  fail "Backend did not become healthy within ${MAX_ATTEMPTS}s"
fi

# ── 9. Start frontend dev server ──────────────────────────────────────

print_step "Starting frontend dev server"
FRONTEND_PID=""
(cd frontend && pnpm dev --port 5173) &
FRONTEND_PID=$!
print_ok "Frontend starting (PID: $FRONTEND_PID)"

FRONTEND_PORT=5173

# Wait for frontend to be ready
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  if curl -s "http://localhost:${FRONTEND_PORT}" >/dev/null 2>&1; then
    print_ok "Frontend is ready at http://localhost:${FRONTEND_PORT}"
    break
  fi
  if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
    fail "Frontend process exited unexpectedly"
  fi
  sleep 1
  ATTEMPT=$((ATTEMPT + 1))
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
  fail "Frontend did not become ready within ${MAX_ATTEMPTS}s"
fi

# ── 10. Open browser ─────────────────────────────────────────────────

FRONTEND_URL="http://localhost:${FRONTEND_PORT}"

STATUS=$(curl -s "http://localhost:${PORT}/admin/status")
INITIALIZED=$(echo "$STATUS" | grep -o '"initialized":[a-z]*' | cut -d: -f2)

if [ "$INITIALIZED" = "true" ]; then
  print_warn "System is already initialized."
else
  print_ok "Opening setup wizard in your browser..."
fi

# Open browser
if [[ "$OSTYPE" == "darwin"* ]]; then
  open "$FRONTEND_URL"
elif command -v xdg-open &>/dev/null; then
  xdg-open "$FRONTEND_URL"
else
  print_warn "Open $FRONTEND_URL in your browser to continue setup."
fi

# ── Done ─────────────────────────────────────────────────────────────

echo ""
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Dev-Share is running!${NC}"
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo ""
echo -e "  API:        http://localhost:${PORT}"
echo -e "  Frontend:   http://localhost:${FRONTEND_PORT}"
echo ""
if [ "$INITIALIZED" != "true" ]; then
  echo -e "  Complete setup in your browser."
  echo ""
fi
echo "Press Ctrl+C to stop."
wait "$BACKEND_PID" "$FRONTEND_PID"
