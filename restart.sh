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

usage() {
  echo "Usage: $0 [backend|frontend|all]"
  echo ""
  echo "  backend   - Restart only the backend server"
  echo "  frontend  - Restart only the frontend dev server"
  echo "  all       - Restart both (default)"
  exit 0
}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

TARGET="${1:-all}"
case "$TARGET" in
  backend|frontend|all) ;;
  -h|--help) usage ;;
  *) fail "Unknown target '$TARGET'. Use: backend, frontend, or all" ;;
esac

# ── 1. Load environment ───────────────────────────────────────────────

print_step "Loading environment"

if [ ! -f .env ]; then
  fail ".env not found. Run setup.sh first."
fi

set -a
source .env
set +a

if [ -z "$JWT_SECRET" ]; then
  fail "JWT_SECRET is not set in .env. Run setup.sh to generate one."
fi
print_ok ".env loaded (JWT_SECRET present)"

# DB_FILE_PATH in .env is relative to the repo root (e.g. ./backend/devshare.db),
# but Go commands run from backend/, so strip the leading ./backend/ prefix.
if [[ "$DB_FILE_PATH" == ./backend/* ]]; then
  export DB_FILE_PATH="./${DB_FILE_PATH#./backend/}"
fi

PORT="${PORT:-8080}"
FRONTEND_PORT=5173
MAX_ATTEMPTS=30

# ── 2. Stop existing processes ────────────────────────────────────────

print_step "Stopping existing processes"

stop_backend() {
  local pids
  pids=$(lsof -ti :"$PORT" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "$pids" | xargs kill 2>/dev/null || true
    sleep 1
    # Force kill if still running
    pids=$(lsof -ti :"$PORT" 2>/dev/null || true)
    if [ -n "$pids" ]; then
      echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
    print_ok "Stopped backend on port $PORT"
  else
    print_warn "No backend process found on port $PORT"
  fi
}

stop_frontend() {
  local pids
  pids=$(lsof -ti :"$FRONTEND_PORT" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "$pids" | xargs kill 2>/dev/null || true
    sleep 1
    pids=$(lsof -ti :"$FRONTEND_PORT" 2>/dev/null || true)
    if [ -n "$pids" ]; then
      echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
    print_ok "Stopped frontend on port $FRONTEND_PORT"
  else
    print_warn "No frontend process found on port $FRONTEND_PORT"
  fi
}

if [ "$TARGET" = "backend" ] || [ "$TARGET" = "all" ]; then
  stop_backend
fi

if [ "$TARGET" = "frontend" ] || [ "$TARGET" = "all" ]; then
  stop_frontend
fi

# ── 3. Cleanup on exit ───────────────────────────────────────────────

BACKEND_PID=""
FRONTEND_PID=""
cleanup() {
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null
  fi
  if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null
  fi
}
trap cleanup EXIT

# ── 4. Start backend ─────────────────────────────────────────────────

if [ "$TARGET" = "backend" ] || [ "$TARGET" = "all" ]; then
  print_step "Starting backend server"
  (cd backend && go run ./cmd/server) &
  BACKEND_PID=$!
  print_ok "Backend starting (PID: $BACKEND_PID)"

  print_step "Waiting for backend to be ready"
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
fi

# ── 5. Start frontend ────────────────────────────────────────────────

if [ "$TARGET" = "frontend" ] || [ "$TARGET" = "all" ]; then
  print_step "Starting frontend dev server"
  (cd frontend && pnpm dev --port "$FRONTEND_PORT") &
  FRONTEND_PID=$!
  print_ok "Frontend starting (PID: $FRONTEND_PID)"

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
fi

# ── Done ──────────────────────────────────────────────────────────────

echo ""
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Dev-Share restarted! ($TARGET)${NC}"
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo ""
if [ "$TARGET" = "backend" ] || [ "$TARGET" = "all" ]; then
  echo -e "  API:        http://localhost:${PORT}"
fi
if [ "$TARGET" = "frontend" ] || [ "$TARGET" = "all" ]; then
  echo -e "  Frontend:   http://localhost:${FRONTEND_PORT}"
fi
echo ""
echo "Press Ctrl+C to stop."

# Wait for running processes
if [ -n "$BACKEND_PID" ] && [ -n "$FRONTEND_PID" ]; then
  wait "$BACKEND_PID" "$FRONTEND_PID"
elif [ -n "$BACKEND_PID" ]; then
  wait "$BACKEND_PID"
elif [ -n "$FRONTEND_PID" ]; then
  wait "$FRONTEND_PID"
fi
