#!/bin/bash
set -euo pipefail

# 1. Update system packages
apt-get update -y
apt-get upgrade -y

# 2. Install Docker Engine
apt-get install -y ca-certificates curl gnupg
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 3. Start and enable Docker service
systemctl start docker
systemctl enable docker

# 4. Add ubuntu user to the docker group
usermod -aG docker ubuntu

# 5. Install git
apt-get install -y git

# 7. Write secret manager config for Dev-Share
# These are NOT actual secrets — just pointers to the AWS Secrets Manager
# secret name and region. The actual secrets are fetched at container
# startup via the Secrets Manager API (see backend/entrypoint.sh).
# Values are injected by Terraform templatefile().
cat > /etc/devshare-secrets.env <<'SECRETS_EOF'
AWS_SECRET_ID=${aws_secret_id}
AWS_DEFAULT_REGION=${aws_region}
SECRETS_EOF
chmod 644 /etc/devshare-secrets.env
