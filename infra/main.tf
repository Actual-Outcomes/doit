terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Shared state bucket with getsit — different key
  backend "s3" {
    bucket         = "akl-terraform-state-632767574247"
    key            = "doit/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "akl-terraform-lock"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "doit"
      ManagedBy = "terraform"
    }
  }
}

# Shared CF2 RDS credentials (same instance as getsit)
data "aws_secretsmanager_secret_version" "cf2_db" {
  secret_id = "cf2-dev-db-credentials"
}
