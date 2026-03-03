#!/usr/bin/env bash
set -euo pipefail

# Create the V&V database on the shared staging RDS instance and
# output Terraform variable values for the V&V environment.
#
# Prerequisites:
#   - AWS CLI configured with appropriate profile
#   - psql installed and accessible
#   - Network access to RDS (VPN or bastion)
#   - jq installed
#
# Usage: ./scripts/setup-vnv-db.sh

AWS_PROFILE="${AWS_PROFILE:-ao-dave}"
AWS_REGION="${AWS_REGION:-us-east-1}"

export AWS_PROFILE AWS_REGION

echo "=== Setting up V&V database ==="

# Get RDS credentials from Secrets Manager
echo "Fetching RDS credentials..."
CREDS=$(aws secretsmanager get-secret-value --secret-id cf2-dev-db-credentials --query 'SecretString' --output text)
DB_HOST=$(echo "$CREDS" | jq -r '.host')
DB_PORT=$(echo "$CREDS" | jq -r '.port // "5432"')
DB_USER=$(echo "$CREDS" | jq -r '.username')
DB_PASS=$(echo "$CREDS" | jq -r '.password')

echo "RDS Host: $DB_HOST"

# Create the V&V database (ignore error if already exists)
echo "Creating database doit_vnv..."
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c \
  "CREATE DATABASE doit_vnv;" 2>/dev/null || echo "Database doit_vnv already exists"

# Generate V&V API key
VNV_API_KEY=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || uuidgen | tr '[:upper:]' '[:lower:]')

# Build the DATABASE_URL
VNV_DATABASE_URL="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/doit_vnv?sslmode=require"

echo ""
echo "=== V&V Database Ready ==="
echo ""
echo "Add these to your terraform apply command:"
echo ""
echo "  terraform apply \\"
echo "    -var 'vnv_database_url=${VNV_DATABASE_URL}' \\"
echo "    -var 'vnv_api_key=${VNV_API_KEY}'"
echo ""
echo "Or export as environment variables:"
echo ""
echo "  export TF_VAR_vnv_database_url='${VNV_DATABASE_URL}'"
echo "  export TF_VAR_vnv_api_key='${VNV_API_KEY}'"
echo ""
echo "V&V API Key (save this): ${VNV_API_KEY}"
