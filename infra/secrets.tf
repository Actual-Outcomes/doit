# Doit-specific application secrets
resource "aws_secretsmanager_secret" "doit_config" {
  name = "doit-config"
}

resource "aws_secretsmanager_secret_version" "doit_config" {
  secret_id = aws_secretsmanager_secret.doit_config.id
  secret_string = jsonencode({
    DATABASE_URL = var.database_url
    API_KEY      = var.api_key
  })
}
