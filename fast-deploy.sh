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

echo "Provision EC2 instance"
CURRENT_DIR=$(pwd)
INFRA_DIR="${CURRENT_DIR}/e2e/infra"
REPO_URL="${REPO_URL:-https://github.com/erel213/dev-share.git}"

# Setup resources
cd "${INFRA_DIR}"
terraform init -input=false
terraform apply -auto-approve -input=false -var "public_key=$(cat "${SSH_KEY}.pub")"
EC2_IP=$(terraform output -raw instance_public_ip)
ok "EC2 instance ready at $EC2_IP"

 
# Clone repo and run setup
ssh $SSH_OPTS "ubuntu@$EC2_IP" bash <<EOF
  set -euo pipefail
  git clone $REPO_URL
  cd dev-share
  git checkout $BRANCH
  ./setup.sh
EOF
echo "Dev-share deployed"

echo "Establish SSH tunnel"
ssh -i "$SSH_KEY" -N -L 3000:localhost:3000 "ubuntu@$EC2_IP" &
echo "SSH tunnel established (localhost:3000 → EC2:3000)"



