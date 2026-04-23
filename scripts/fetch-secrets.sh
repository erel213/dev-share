#!/bin/bash
# Fetches Dev-Share application secrets from a cloud secret manager on the host
# and writes them to a tmpfs-backed directory for the backend container to consume
# via the *_FILE env var convention (see docker-compose.cloud-secrets.yml).
#
# Usage: set one of AWS_SECRET_ID / AZURE_KEYVAULT_NAME / GCP_SECRET_NAME, then run.
# Optional: SECRETS_DIR (default: /dev/shm/devshare-secrets).
#
# On success, prints `SECRETS_DIR=<path>` to stdout so callers can `eval` it.
# All status/error messages go to stderr.

set -euo pipefail
umask 077

log()  { echo "[fetch-secrets] $*" >&2; }
fail() { echo "[fetch-secrets] error: $*" >&2; exit 1; }

# Tmpfs only exists on Linux. macOS dev should use .env (Option A).
if [ "$(uname -s)" != "Linux" ]; then
  fail "host OS is $(uname -s); tmpfs-backed secret flow is Linux-only. Use a local .env instead (./setup.sh --generate-env)."
fi

SECRETS_DIR="${SECRETS_DIR:-/dev/shm/devshare-secrets}"
mkdir -p "$SECRETS_DIR"
chmod 700 "$SECRETS_DIR"

# Always create all three files so compose `secrets:` mounts never fail on a
# missing source. ADMIN_INIT_TOKEN is optional and may remain empty.
: > "$SECRETS_DIR/jwt_secret"
: > "$SECRETS_DIR/encryption_key"
: > "$SECRETS_DIR/admin_init_token"
chmod 600 "$SECRETS_DIR"/{jwt_secret,encryption_key,admin_init_token}

require_cli() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required on the host but was not found in PATH"
}

# jq is used to parse JSON payloads from AWS/GCP. Azure stores each value as a
# separate Key Vault secret and does not need jq.
write_from_json() {
  local json="$1"
  command -v jq >/dev/null 2>&1 || fail "jq is required on the host to parse the secret JSON payload"
  jq -r '.JWT_SECRET // ""'        <<<"$json" > "$SECRETS_DIR/jwt_secret"
  jq -r '.ENCRYPTION_KEY // ""'    <<<"$json" > "$SECRETS_DIR/encryption_key"
  jq -r '.ADMIN_INIT_TOKEN // ""'  <<<"$json" > "$SECRETS_DIR/admin_init_token"
}

if [ -n "${AWS_SECRET_ID:-}" ]; then
  log "fetching from AWS Secrets Manager (AWS_SECRET_ID=$AWS_SECRET_ID)"
  require_cli aws
  SECRET_JSON="$(aws secretsmanager get-secret-value \
    --secret-id "$AWS_SECRET_ID" \
    --query SecretString \
    --output text)"
  write_from_json "$SECRET_JSON"

elif [ -n "${AZURE_KEYVAULT_NAME:-}" ]; then
  log "fetching from Azure Key Vault (AZURE_KEYVAULT_NAME=$AZURE_KEYVAULT_NAME)"
  require_cli az
  az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name jwt-secret       --query value -o tsv > "$SECRETS_DIR/jwt_secret"
  az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name encryption-key   --query value -o tsv > "$SECRETS_DIR/encryption_key"
  az keyvault secret show --vault-name "$AZURE_KEYVAULT_NAME" --name admin-init-token --query value -o tsv > "$SECRETS_DIR/admin_init_token" 2>/dev/null || true

elif [ -n "${GCP_SECRET_NAME:-}" ]; then
  log "fetching from GCP Secret Manager (GCP_SECRET_NAME=$GCP_SECRET_NAME)"
  require_cli gcloud
  SECRET_JSON="$(gcloud secrets versions access latest --secret="$GCP_SECRET_NAME")"
  write_from_json "$SECRET_JSON"

else
  fail "no cloud provider ID set — export one of AWS_SECRET_ID, AZURE_KEYVAULT_NAME, GCP_SECRET_NAME"
fi

# Minimal sanity check — don't leak values, just lengths.
jwt_len=$(wc -c < "$SECRETS_DIR/jwt_secret" | tr -d ' ')
enc_len=$(wc -c < "$SECRETS_DIR/encryption_key" | tr -d ' ')
log "wrote jwt_secret (${jwt_len} bytes), encryption_key (${enc_len} bytes) to $SECRETS_DIR"

echo "SECRETS_DIR=$SECRETS_DIR"
