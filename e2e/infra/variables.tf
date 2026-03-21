variable "region" {
  description = "AWS region to deploy the dev-share host"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type (needs enough resources for Docker)"
  type        = string
  default     = "t3.medium"
}

variable "key_name" {
  description = "Name of the SSH key pair to create for EC2 access"
  type        = string
  default     = "dev-share-e2e"
}

variable "public_key_path" {
  description = "Path to the SSH public key file to import"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}
