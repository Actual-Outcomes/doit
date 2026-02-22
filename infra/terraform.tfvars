# Shared with getsit — same VPC, subnets, RDS
aws_region = "us-east-1"

vpc_id = "vpc-0250c9385a5501236"

private_subnet_ids = [
  "subnet-0b666f6464538246a",
  "subnet-072443afbc98b6d59",
]

db_security_group_id = "sg-0ec329aee45c23907"

rds_endpoint = "cf2-data-dev-stack-postgresinstance19cdd68a-bde0nkq5pwrg.cbo9sxcxkohm.us-east-1.rds.amazonaws.com"
