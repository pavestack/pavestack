# VPC module

Creates a production-aligned VPC with public and private subnets across multiple AZs and a single NAT gateway for cost-efficient egress from private subnets. VPC flow logs are enabled by default, delivered to a KMS-encrypted CloudWatch Logs group.

## Inputs

- `enable_flow_logs` - enable VPC flow logs delivered to a KMS-encrypted CloudWatch Logs group (default `true`)
- `flow_log_retention_days` - CloudWatch Logs retention in days for VPC flow logs (default `30`)

## Outputs

- `vpc_id`
- `vpc_cidr`
- `public_subnet_ids`
- `private_subnet_ids`
- `availability_zones`
- `flow_log_id`
- `flow_log_group_name`
- `flow_log_group_arn`
