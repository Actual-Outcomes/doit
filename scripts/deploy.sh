#!/usr/bin/env bash
set -euo pipefail

# Deploy doit to AWS App Runner
# Usage: ./scripts/deploy.sh [--profile PROFILE] [--tag TAG]

AWS_PROFILE="${AWS_PROFILE:-ao-dave}"
AWS_REGION="${AWS_REGION:-us-east-1}"
IMAGE_TAG="latest"

while [[ $# -gt 0 ]]; do
  case $1 in
    --profile) AWS_PROFILE="$2"; shift 2 ;;
    --tag)     IMAGE_TAG="$2"; shift 2 ;;
    *)         echo "Unknown option: $1"; exit 1 ;;
  esac
done

export AWS_PROFILE AWS_REGION

echo "=== Deploying doit (profile=$AWS_PROFILE, tag=$IMAGE_TAG) ==="

# Read Terraform outputs
cd "$(dirname "$0")/../infra"
ECR_URL=$(terraform output -raw ecr_repository_url)
SERVICE_ARN=$(terraform output -raw service_arn)
ACCOUNT_ID=$(echo "$ECR_URL" | cut -d. -f1)
cd ..

echo "ECR: $ECR_URL"
echo "Service: $SERVICE_ARN"

# Docker login to ECR
echo "=== Logging into ECR ==="
aws ecr get-login-password --region "$AWS_REGION" | \
  docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

# Resolve version from git tags
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
echo "Version: $VERSION"

# Build and push
echo "=== Building Docker image ==="
docker build --build-arg "VERSION=$VERSION" -t "doit:$IMAGE_TAG" -f Dockerfile .

echo "=== Pushing to ECR ==="
docker tag "doit:$IMAGE_TAG" "$ECR_URL:$IMAGE_TAG"
docker push "$ECR_URL:$IMAGE_TAG"

# Deploy via App Runner
echo "=== Starting deployment ==="
aws apprunner start-deployment --service-arn "$SERVICE_ARN" --region "$AWS_REGION"

echo ""
echo "=== Deployment initiated ==="
echo "Console: https://$AWS_REGION.console.aws.amazon.com/apprunner/home#/services"
echo "Check status: aws apprunner describe-service --service-arn $SERVICE_ARN --query 'Service.Status' --output text"
