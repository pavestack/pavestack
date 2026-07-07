# vpc module

Creates a production-aligned VPC with public and private subnets spread across multiple availability zones, plus NAT gateway egress for private subnets. VPC flow logs are enabled by default and delivered to a KMS-encrypted CloudWatch Logs group.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 6.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 6.0 |

## Resources

| Name | Type |
|------|------|
| aws_availability_zones.available | data source |
| aws_caller_identity.current | data source |
| aws_cloudwatch_log_group.flow_logs | resource |
| aws_default_security_group.default | resource |
| aws_eip.nat | resource |
| aws_flow_log.this | resource |
| aws_iam_policy_document.flow_logs | data source |
| aws_iam_policy_document.flow_logs_assume_role | data source |
| aws_iam_role.flow_logs | resource |
| aws_iam_role_policy.flow_logs | resource |
| aws_internet_gateway.this | resource |
| aws_kms_alias.flow_logs | resource |
| aws_kms_key.flow_logs | resource |
| aws_nat_gateway.this | resource |
| aws_region.current | data source |
| aws_route_table.private | resource |
| aws_route_table.public | resource |
| aws_route_table_association.private | resource |
| aws_route_table_association.public | resource |
| aws_subnet.private | resource |
| aws_subnet.public | resource |
| aws_vpc.this | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_az_count"></a> [az\_count](#input\_az\_count) | Number of availability zones to use. | `number` | `3` | no |
| <a name="input_enable_flow_logs"></a> [enable\_flow\_logs](#input\_enable\_flow\_logs) | Enable VPC flow logs delivered to a KMS-encrypted CloudWatch Logs group. | `bool` | `true` | no |
| <a name="input_enable_nat_gateway"></a> [enable\_nat\_gateway](#input\_enable\_nat\_gateway) | Create NAT gateways for private subnet egress. | `bool` | `true` | no |
| <a name="input_flow_log_retention_days"></a> [flow\_log\_retention\_days](#input\_flow\_log\_retention\_days) | CloudWatch Logs retention in days for VPC flow logs. Must be one of the values accepted by aws\_cloudwatch\_log\_group (e.g. 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653, or 0 for never expire). | `number` | `30` | no |
| <a name="input_name"></a> [name](#input\_name) | Name prefix for VPC resources. | `string` | n/a | yes |
| <a name="input_single_nat_gateway"></a> [single\_nat\_gateway](#input\_single\_nat\_gateway) | Use one NAT gateway for all private subnets. | `bool` | `false` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |
| <a name="input_vpc_cidr"></a> [vpc\_cidr](#input\_vpc\_cidr) | CIDR block for the VPC. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_availability_zones"></a> [availability\_zones](#output\_availability\_zones) | Selected availability zones. |
| <a name="output_flow_log_group_arn"></a> [flow\_log\_group\_arn](#output\_flow\_log\_group\_arn) | CloudWatch Logs group ARN for VPC flow logs (null when flow logs are disabled). |
| <a name="output_flow_log_group_name"></a> [flow\_log\_group\_name](#output\_flow\_log\_group\_name) | CloudWatch Logs group name for VPC flow logs (null when flow logs are disabled). |
| <a name="output_flow_log_id"></a> [flow\_log\_id](#output\_flow\_log\_id) | VPC flow log ID (null when flow logs are disabled). |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | Private subnet IDs. |
| <a name="output_public_subnet_ids"></a> [public\_subnet\_ids](#output\_public\_subnet\_ids) | Public subnet IDs. |
| <a name="output_vpc_cidr"></a> [vpc\_cidr](#output\_vpc\_cidr) | VPC CIDR block. |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | VPC ID. |
<!-- END_TF_DOCS -->
