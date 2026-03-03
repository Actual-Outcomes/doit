output "service_url" {
  description = "HTTPS URL of the doit App Runner service"
  value       = "https://${aws_apprunner_service.doit.service_url}"
}

output "ecr_repository_url" {
  description = "ECR repository URL for pushing doit images"
  value       = aws_ecr_repository.doit.repository_url
}

output "service_arn" {
  description = "App Runner service ARN for deployments"
  value       = aws_apprunner_service.doit.arn
}

# V&V environment
output "vnv_service_url" {
  description = "HTTPS URL of the doit V&V App Runner service"
  value       = "https://${aws_apprunner_service.doit_vnv.service_url}"
}

output "vnv_service_arn" {
  description = "App Runner service ARN for V&V deployments"
  value       = aws_apprunner_service.doit_vnv.arn
}
