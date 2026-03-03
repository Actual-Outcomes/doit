# V&V (Verification & Validation) environment
# Reuses: ECR repo, VPC connector, security groups
# Separate: App Runner service, secrets, IAM instance role

# --- Secrets ---

resource "aws_secretsmanager_secret" "doit_vnv_config" {
  name = "doit-vnv-config"
}

resource "aws_secretsmanager_secret_version" "doit_vnv_config" {
  secret_id = aws_secretsmanager_secret.doit_vnv_config.id
  secret_string = jsonencode({
    DATABASE_URL = var.vnv_database_url
    API_KEY      = var.vnv_api_key
  })
}

# --- IAM (least privilege — only V&V secrets) ---

resource "aws_iam_role" "apprunner_vnv_instance" {
  name = "doit-vnv-apprunner-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "tasks.apprunner.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "apprunner_vnv_secrets" {
  name = "doit-vnv-secrets-access"
  role = aws_iam_role.apprunner_vnv_instance.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.doit_vnv_config.arn
        ]
      }
    ]
  })
}

# --- App Runner Service ---

resource "aws_apprunner_service" "doit_vnv" {
  service_name = "doit-vnv"

  source_configuration {
    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_ecr.arn
    }

    image_repository {
      image_identifier      = "${aws_ecr_repository.doit.repository_url}:latest"
      image_repository_type = "ECR"

      image_configuration {
        port = "8080"
        runtime_environment_secrets = {
          DATABASE_URL = "${aws_secretsmanager_secret.doit_vnv_config.arn}:DATABASE_URL::"
          API_KEY      = "${aws_secretsmanager_secret.doit_vnv_config.arn}:API_KEY::"
        }
        runtime_environment_variables = {
          LOG_LEVEL         = "info"
          ID_PREFIX         = "doit"
          PORT              = "8080"
          ADMIN_TENANT_SLUG = "actual-outcomes"
        }
      }
    }

    auto_deployments_enabled = false
  }

  network_configuration {
    egress_configuration {
      egress_type       = "VPC"
      vpc_connector_arn = aws_apprunner_vpc_connector.doit.arn
    }
  }

  instance_configuration {
    cpu               = "256"
    memory            = "512"
    instance_role_arn = aws_iam_role.apprunner_vnv_instance.arn
  }

  health_check_configuration {
    protocol            = "HTTP"
    path                = "/health"
    interval            = 10
    timeout             = 5
    healthy_threshold   = 1
    unhealthy_threshold = 3
  }
}
