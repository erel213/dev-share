#!/bin/bash
# This script will provision an EC2 instance using Terraform and deploy the application on it.
# It requires the user to provide the path to their SSH private key, which will be used to access the EC2 instance after it is provisioned.
# Usage: ./fast-deploy.sh --ssh-key <path-to-ssh-private-key>

# Parse command line arguments
SSH_KEY=""
BRANCH="main"
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --ssh-key) SSH_KEY="$2"; shift ;;
        --branch) BRANCH="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$SSH_KEY" ]; then
  echo "Usage: $0 --ssh-key <path-to-ssh-private-key> [--branch <branch>]"
  exit 1
fi

SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"

echo "Provision EC2 instance"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRA_DIR="${SCRIPT_DIR}/e2e/infra"
OVERRIDE_FILE="${SCRIPT_DIR}/e2e/docker-compose.override.e2e.yml"
REPO_URL="${REPO_URL:-https://github.com/erel213/dev-share.git}"

# Setup resources
cd "${INFRA_DIR}"
terraform init -input=false
terraform apply -auto-approve -input=false -var "public_key=$(cat "${SSH_KEY}.pub")"
EC2_IP=$(terraform output -raw instance_public_ip)
echo "EC2 instance ready at $EC2_IP"

# Wait for SSH, then for cloud-init so /etc/devshare-secrets.env (written by user-data) exists
until ssh $SSH_OPTS -o ConnectTimeout=5 "ubuntu@$EC2_IP" "echo ready" 2>/dev/null; do
  sleep 5
done
ssh $SSH_OPTS "ubuntu@$EC2_IP" "cloud-init status --wait" 2>/dev/null

# setup.sh's cloud-secret branch requires docker-compose.override.yml with JWT_SECRET_FILE
scp $SSH_OPTS "$OVERRIDE_FILE" "ubuntu@$EC2_IP:/tmp/docker-compose.override.yml"

# Source AWS_SECRET_ID + AWS_DEFAULT_REGION so setup.sh takes the cloud-secret path
ssh $SSH_OPTS "ubuntu@$EC2_IP" bash -l <<EOF
  set -euo pipefail
  git clone $REPO_URL
  cd dev-share
  git checkout $BRANCH
  cp /tmp/docker-compose.override.yml ./docker-compose.override.yml
  ./setup.sh
EOF
echo "Dev-share deployed"

echo "Establish SSH tunnel"
ssh $SSH_OPTS -N -L 3000:localhost:3000 "ubuntu@$EC2_IP" &
echo "SSH tunnel established (localhost:3000 → EC2:3000)"



