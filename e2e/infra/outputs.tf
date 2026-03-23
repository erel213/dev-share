output "instance_public_ip" {
  description = "Public IP of the dev-share E2E host"
  value       = aws_instance.e2e.public_ip
}

output "instance_id" {
  description = "Instance ID of the dev-share E2E host"
  value       = aws_instance.e2e.id
}
