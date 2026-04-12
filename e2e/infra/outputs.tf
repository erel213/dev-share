output "instance_public_ip" {
  description = "Public IP of the dev-share E2E host"
  value       = aws_instance.e2e.public_ip
}

output "instance_id" {
  description = "Instance ID of the dev-share E2E host"
  value       = aws_instance.e2e.id
}

output "iam_role_arn" {
  description = "ARN of the IAM role attached to the E2E instance"
  value       = aws_iam_role.e2e.arn
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret for Dev-Share"
  value       = aws_secretsmanager_secret.devshare.arn
}

output "secret_name" {
  description = "Name of the Secrets Manager secret (use as AWS_SECRET_ID)"
  value       = aws_secretsmanager_secret.devshare.name
}
