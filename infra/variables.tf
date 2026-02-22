variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# Shared networking — same VPC and subnets as getsit/CF2
variable "vpc_id" {
  description = "VPC ID (CF2 dev VPC, shared with getsit)"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for VPC Connector (shared with getsit)"
  type        = list(string)
}

# Shared RDS — same instance as getsit, different database
variable "db_security_group_id" {
  description = "Security group ID of the shared CF2 RDS instance"
  type        = string
}

variable "rds_endpoint" {
  description = "Shared RDS endpoint hostname"
  type        = string
}

# Doit-specific secrets
variable "database_url" {
  description = "PostgreSQL connection string for the doit database"
  type        = string
  sensitive   = true
}

variable "api_key" {
  description = "Admin API key for doit"
  type        = string
  sensitive   = true
}
