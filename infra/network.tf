# Doit's own security group for App Runner
resource "aws_security_group" "doit_apprunner" {
  name        = "doit-apprunner-sg"
  description = "Security group for doit App Runner VPC connector"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }
}

# Allow doit App Runner to reach the shared RDS
resource "aws_vpc_security_group_ingress_rule" "doit_to_rds" {
  security_group_id            = var.db_security_group_id
  referenced_security_group_id = aws_security_group.doit_apprunner.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
  description                  = "doit App Runner to shared RDS"
}

# VPC Connector for doit App Runner → private subnets
resource "aws_apprunner_vpc_connector" "doit" {
  vpc_connector_name = "doit-vpc-connector"
  subnets            = var.private_subnet_ids
  security_groups    = [aws_security_group.doit_apprunner.id]
}

# App Runner service for doit
resource "aws_apprunner_service" "doit" {
  service_name = "doit"

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
          DATABASE_URL = "${aws_secretsmanager_secret.doit_config.arn}:DATABASE_URL::"
          API_KEY      = "${aws_secretsmanager_secret.doit_config.arn}:API_KEY::"
        }
        runtime_environment_variables = {
          LOG_LEVEL         = "info"
          ID_PREFIX         = "doit"
          PORT              = "8080"
          ADMIN_TENANT_SLUG = "default"
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
    instance_role_arn = aws_iam_role.apprunner_instance.arn
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

# IAM role for App Runner to pull from ECR
resource "aws_iam_role" "apprunner_ecr" {
  name = "doit-apprunner-ecr-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "build.apprunner.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "apprunner_ecr" {
  role       = aws_iam_role.apprunner_ecr.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAppRunnerServicePolicyForECRAccess"
}

# App Runner instance role for Secrets Manager access
resource "aws_iam_role" "apprunner_instance" {
  name = "doit-apprunner-instance-role"

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

resource "aws_iam_role_policy" "apprunner_secrets" {
  name = "doit-secrets-access"
  role = aws_iam_role.apprunner_instance.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.doit_config.arn
        ]
      }
    ]
  })
}
