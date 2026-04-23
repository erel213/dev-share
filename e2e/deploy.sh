#!/bin/bash
set -euo pipefail

# ── Argument parsing ─────────────────────────────────────────────────
EXPORT_REPORT=false
for arg in "$@"; do
  case "$arg" in
    --export-report) EXPORT_REPORT=true ;;
  esac
done

# ── Configuration ─────────────────────────────────────────────────────
CALLER_DIR="$(pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRA_DIR="$SCRIPT_DIR/infra"
SSH_KEY=$(mktemp /tmp/e2e-key-XXXXX)
rm -f "$SSH_KEY"
ssh-keygen -t ed25519 -f "$SSH_KEY" -N "" -q
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"
REPO_URL="${REPO_URL:-https://github.com/erel213/dev-share.git}"
SSH_TIMEOUT=300  # 5 minutes max wait for SSH
HEALTH_TIMEOUT=120  # 2 minutes max wait for app health

# ── Colors ────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${BLUE}${BOLD}==>${NC}${BOLD} $1${NC}"; }
ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }
err()  { echo -e "  ${RED}✗${NC} $1"; }

# ── Cleanup (runs on EXIT, regardless of success/failure) ─────────────
cleanup() {
  echo ""
  log "Tearing down..."

  # Kill SSH tunnel
  if [ -n "${TUNNEL_PID:-}" ] && kill -0 "$TUNNEL_PID" 2>/dev/null; then
    kill "$TUNNEL_PID" 2>/dev/null
    ok "SSH tunnel stopped"
  fi

  # Remove ephemeral SSH key
  if [ -n "${SSH_KEY:-}" ]; then
    rm -f "$SSH_KEY" "${SSH_KEY}.pub"
    ok "Ephemeral SSH key removed"
  fi

  # Destroy infrastructure
  if [ -d "$INFRA_DIR/.terraform" ]; then
    cd "$INFRA_DIR"
    terraform destroy -auto-approve 2>/dev/null && ok "Infrastructure destroyed" || warn "Terraform destroy failed"
  fi
}
trap cleanup EXIT

# ── Step 1: Provision EC2 ─────────────────────────────────────────────
log "Provisioning EC2 instance"
cd "$INFRA_DIR"
terraform init -input=false
terraform apply -auto-approve -input=false -var "public_key=$(cat "${SSH_KEY}.pub")"
EC2_IP=$(terraform output -raw instance_public_ip)
ok "EC2 instance ready at $EC2_IP"

# ── Step 2: Wait for SSH ─────────────────────────────────────────────
log "Waiting for SSH to become available (timeout: ${SSH_TIMEOUT}s)"
SECONDS=0
until ssh $SSH_OPTS -o ConnectTimeout=5 "ubuntu@$EC2_IP" "echo ready" 2>/dev/null; do
  if [ $SECONDS -ge $SSH_TIMEOUT ]; then
    err "SSH did not become available within ${SSH_TIMEOUT}s"
    exit 1
  fi
  sleep 5
done
ok "SSH is available (${SECONDS}s)"

# ── Step 3: Wait for cloud-init to finish ─────────────────────────────
log "Waiting for cloud-init to complete"
ssh $SSH_OPTS "ubuntu@$EC2_IP" "cloud-init status --wait" 2>/dev/null
ok "Cloud-init completed"

# ── Step 4: Clone repo and run setup ──────────────────────────────────
log "Deploying dev-share on EC2"
scp $SSH_OPTS "$SCRIPT_DIR/docker-compose.override.e2e.yml" "ubuntu@$EC2_IP:/tmp/docker-compose.override.yml"
ssh $SSH_OPTS "ubuntu@$EC2_IP" bash <<EOF
  set -euo pipefail
  git clone $REPO_URL
  cd dev-share
  cp /tmp/docker-compose.override.yml ./docker-compose.override.yml
  set -a; . /etc/devshare-secrets.env; set +a
  ./setup.sh
EOF
ok "Dev-share deployed"

# ── Step 5: Establish SSH tunnel ──────────────────────────────────────
log "Establishing SSH tunnel (localhost:3000 → EC2:3000)"
ssh $SSH_OPTS -L 3000:localhost:3000 -N -f "ubuntu@$EC2_IP"
TUNNEL_PID=$(lsof -ti:3000 2>/dev/null | head -1)
ok "SSH tunnel established (PID: ${TUNNEL_PID:-unknown})"

# ── Step 6: Verify app is reachable through tunnel ────────────────────
log "Verifying app is reachable through tunnel (timeout: ${HEALTH_TIMEOUT}s)"
SECONDS=0
until curl -s http://localhost:3000/health >/dev/null 2>&1; do
  if [ $SECONDS -ge $HEALTH_TIMEOUT ]; then
    err "App not reachable through tunnel within ${HEALTH_TIMEOUT}s"
    exit 1
  fi
  sleep 2
done
ok "App is healthy at http://localhost:3000 (${SECONDS}s)"

# ── Step 7: Install Playwright dependencies ─────────────────────────
log "Installing Playwright dependencies"
cd "$SCRIPT_DIR"
npm ci
npx playwright install --with-deps chromium
ok "Playwright ready"

# ── Step 8: Run Playwright tests ──────────────────────────────────────
log "Running Playwright tests"
npx playwright test
TEST_EXIT=$?

if [ $TEST_EXIT -eq 0 ]; then
  ok "All tests passed"
else
  err "Tests failed with exit code $TEST_EXIT"
fi

# ── Step 9: Export Playwright report ──────────────────────────────────
if [ "$EXPORT_REPORT" = true ]; then
  log "Exporting Playwright report"
  REPORT_DIR="$SCRIPT_DIR/playwright-report"
  if [ -d "$REPORT_DIR" ]; then
    cp -r "$REPORT_DIR" "$CALLER_DIR/playwright-report"
    ok "Report exported to $CALLER_DIR/playwright-report"
  else
    warn "No Playwright report found at $REPORT_DIR"
  fi
fi

exit $TEST_EXIT
