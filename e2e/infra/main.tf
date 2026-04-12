terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "aws" {
  region = var.region
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_key_pair" "e2e" {
  key_name   = var.key_name
  public_key = var.public_key
}

resource "aws_security_group" "e2e" {
  name        = "dev-share-e2e-sg"
  description = "Allow SSH inbound only for E2E testing"

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "dev-share-e2e"
  }
}

# --- Secrets ---

resource "random_password" "jwt_secret" {
  length  = 44
  special = false
}

resource "random_id" "encryption_key" {
  byte_length = 32
}

resource "aws_secretsmanager_secret" "devshare" {
  name                    = "devshare/app-secrets"
  description             = "Dev-Share application secrets (JWT + encryption key)"
  recovery_window_in_days = 0 # E2E: allow immediate deletion on terraform destroy

  tags = {
    Name = "dev-share-e2e"
  }
}

resource "aws_secretsmanager_secret_version" "devshare" {
  secret_id = aws_secretsmanager_secret.devshare.id
  secret_string = jsonencode({
    JWT_SECRET     = random_password.jwt_secret.result
    ENCRYPTION_KEY = random_id.encryption_key.hex
  })
}

# --- IAM ---

resource "aws_iam_role" "e2e" {
  name = "dev-share-e2e-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })

  tags = {
    Name = "dev-share-e2e"
  }
}

resource "aws_iam_instance_profile" "e2e" {
  name = "dev-share-e2e-profile"
  role = aws_iam_role.e2e.name
}

resource "aws_iam_role_policy" "secrets_access" {
  name = "dev-share-secrets-access"
  role = aws_iam_role.e2e.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ]
      Resource = aws_secretsmanager_secret.devshare.arn
    }]
  })
}

resource "aws_iam_role_policy_attachment" "power_user" {
  role       = aws_iam_role.e2e.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

# --- EC2 Instance ---

resource "aws_instance" "e2e" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = aws_key_pair.e2e.key_name
  vpc_security_group_ids = [aws_security_group.e2e.id]
  iam_instance_profile   = aws_iam_instance_profile.e2e.name
  user_data = templatefile("${path.module}/user-data.sh", {
    aws_secret_id = aws_secretsmanager_secret.devshare.name
    aws_region    = var.region
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = {
    Name = "dev-share-e2e"
  }
}
