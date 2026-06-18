# VPC module

Creates a production-aligned VPC with public and private subnets across multiple AZs and a single NAT gateway for cost-efficient egress from private subnets.

## Outputs

- `vpc_id`
- `vpc_cidr`
- `public_subnet_ids`
- `private_subnet_ids`
- `availability_zones`
