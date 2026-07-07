output "vpc_id" {
  description = "VPC ID."
  value       = aws_vpc.this.id
}

output "vpc_cidr" {
  description = "VPC CIDR block."
  value       = aws_vpc.this.cidr_block
}

output "public_subnet_ids" {
  description = "Public subnet IDs."
  value       = values(aws_subnet.public)[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs."
  value       = values(aws_subnet.private)[*].id
}

output "availability_zones" {
  description = "Selected availability zones."
  value       = local.azs
}

output "flow_log_id" {
  description = "VPC flow log ID (null when flow logs are disabled)."
  value       = try(aws_flow_log.this[0].id, null)
}

output "flow_log_group_name" {
  description = "CloudWatch Logs group name for VPC flow logs (null when flow logs are disabled)."
  value       = try(aws_cloudwatch_log_group.flow_logs[0].name, null)
}

output "flow_log_group_arn" {
  description = "CloudWatch Logs group ARN for VPC flow logs (null when flow logs are disabled)."
  value       = try(aws_cloudwatch_log_group.flow_logs[0].arn, null)
}

