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

# 6. Install AWS CLI v2
apt-get install -y unzip
arch=$(uname -m)
case "$arch" in
  x86_64) awscli_url="https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" ;;
  aarch64) awscli_url="https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac
tmpdir=$(mktemp -d)
curl -fsSL "$awscli_url" -o "$tmpdir/awscliv2.zip"
unzip -q "$tmpdir/awscliv2.zip" -d "$tmpdir"
"$tmpdir/aws/install"
rm -rf "$tmpdir"

export AWS_SECRET_ID=${aws_secret_id}
export AWS_DEFAULT_REGION=${aws_region}
