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

GENERATE_ENV=false
for arg in "$@"; do
  case "$arg" in
    --generate-env) GENERATE_ENV=true ;;
  esac
done

cleanup() {
  if [ -n "$COMPOSE_UP" ]; then
    echo ""
    print_step "Stopping Dev-Share..."
    docker compose down
  fi
  if [ -n "${SECRETS_DIR:-}" ] && [ -d "$SECRETS_DIR" ]; then
    shred -u "$SECRETS_DIR"/* 2>/dev/null || rm -f "$SECRETS_DIR"/*
    rmdir "$SECRETS_DIR" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# ── 1. Check prerequisites ──────────────────────────────────────────

print_step "Checking prerequisites"

if ! command -v docker &>/dev/null; then
  fail "Docker is not installed. Please install Docker from https://docs.docker.com/get-docker/"
fi
print_ok "Docker $(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')"

if ! docker compose version &>/dev/null; then
  fail "Docker Compose is not available. Please install Docker Compose: https://docs.docker.com/compose/install/"
fi
print_ok "Docker Compose $(docker compose version --short)"

if ! docker info &>/dev/null 2>&1; then
  fail "Docker daemon is not running. Please start Docker and try again."
fi
print_ok "Docker daemon is running"

if ! command -v curl &>/dev/null; then
  fail "curl is not installed"
fi
print_ok "curl"

# ── 2. Set up environment ───────────────────────────────────────────

print_step "Setting up environment"

if [ -f .env ]; then
  if [ "$GENERATE_ENV" = "true" ]; then
    print_warn ".env already exists, skipping generation"
  else
    print_ok "Using existing .env"
  fi
elif [ "$GENERATE_ENV" = "true" ]; then
  print_ok "Generating .env with random secrets"
  cp .env.example .env
  JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
  ENCRYPTION_KEY=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p -c 32)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    sed -i '' "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
  else
    sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    sed -i "s|^ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
  fi
  print_ok ".env created with generated secrets"
else
  if [ -n "${AWS_SECRET_ID:-}" ] || [ -n "${AZURE_KEYVAULT_NAME:-}" ] || [ -n "${GCP_SECRET_NAME:-}" ]; then
    if [ ! -f docker-compose.override.yml ] || ! grep -q '^[^#]*JWT_SECRET_FILE' docker-compose.override.yml; then
      fail "Cloud-secret flow requires docker-compose.override.yml with OPTION A uncommented. Run: cp docker-compose.override.example.yml docker-compose.override.yml && edit to uncomment section A (and the top-level secrets: block)."
    fi
    print_warn "No .env file found — fetching secrets from cloud secret manager on the host"
    eval "$(./scripts/fetch-secrets.sh)"
    export SECRETS_DIR
    print_ok "Secrets written to tmpfs at $SECRETS_DIR (mounted into container as /run/secrets/*)"
  else
    fail "No .env found and no cloud secret manager ID set. Either run './setup.sh --generate-env' or export one of AWS_SECRET_ID / AZURE_KEYVAULT_NAME / GCP_SECRET_NAME."
  fi
fi

# ── 3. Build and start containers ────────────────────────────────────

print_step "Building and starting containers"
if ! docker compose up --build -d; then
  print_error "Failed to start containers. Dumping logs:"
  docker compose logs
  fail "Failed to start containers"
fi
COMPOSE_UP=1
print_ok "Containers started"

# ── 4. Wait for healthy ─────────────────────────────────────────────

print_step "Waiting for Dev-Share to be ready"

APP_PORT="${APP_PORT:-3000}"
MAX_ATTEMPTS=60
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  if curl -s "http://localhost:${APP_PORT}/health" >/dev/null 2>&1; then
    print_ok "Dev-Share is healthy at http://localhost:${APP_PORT}"
    break
  fi
  sleep 1
  ATTEMPT=$((ATTEMPT + 1))
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
  echo ""
  print_error "Dev-Share did not become healthy within ${MAX_ATTEMPTS}s"
  print_warn "Container logs:"
  docker compose logs
  exit 1
fi

print_step "Backend startup logs"
docker compose logs backend

# ── 5. Open browser ─────────────────────────────────────────────────

APP_URL="http://localhost:${APP_PORT}/setup"

STATUS=$(curl -s "http://localhost:${APP_PORT}/admin/status")
INITIALIZED=$(echo "$STATUS" | grep -o '"initialized":[a-z]*' | cut -d: -f2)

if [ "$INITIALIZED" = "true" ]; then
  print_warn "System is already initialized."
  APP_URL="http://localhost:${APP_PORT}"
else
  print_ok "Opening setup wizard in your browser..."
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
  open "$APP_URL"
elif command -v xdg-open &>/dev/null; then
  xdg-open "$APP_URL"
else
  print_warn "Open $APP_URL in your browser to continue setup."
fi

# ── Done ─────────────────────────────────────────────────────────────

echo ""
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Dev-Share is running!${NC}"
echo -e "${GREEN}${BOLD}════════════════════════════════════════${NC}"
echo ""
echo -e "  App:   http://localhost:${APP_PORT}"
echo ""
if [ "$INITIALIZED" != "true" ]; then
  echo -e "  Complete setup in your browser."
  echo ""
fi
echo -e "  Logs:  docker compose logs -f"
echo -e "  Stop:  docker compose down"
echo ""
COMPOSE_UP=""
