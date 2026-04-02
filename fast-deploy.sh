#!/bin/bash
# This script will provision an EC2 instance using Terraform and deploy the application on it.
# It requires the user to provide the path to their SSH private key, which will be used to access the EC2 instance after it is provisioned.
# Usage: ./fast-deploy.sh --ssh-key <path-to-ssh-private-key>

# Parse ssh private key path from command line argument
SSH_KEY=""
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --ssh-key) SSH_KEY="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

echo "Provision EC2 instance"
CURRENT_DIR=$(pwd)
INFRA_DIR="${CURRENT_DIR}/e2e/infra"

cd "${INFRA_DIR}"
terraform init -input=false
terraform apply -auto-approve -input=false -var "public_key=$(cat "${SSH_KEY}.pub")"



