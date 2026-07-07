variable "name" {
  description = "Name prefix for VPC resources."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
}

variable "az_count" {
  description = "Number of availability zones to use."
  type        = number
  default     = 3
}

variable "enable_nat_gateway" {
  description = "Create NAT gateways for private subnet egress."
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use one NAT gateway for all private subnets."
  type        = bool
  default     = false
}

variable "enable_flow_logs" {
  description = "Enable VPC flow logs delivered to a KMS-encrypted CloudWatch Logs group."
  type        = bool
  default     = true
}

variable "flow_log_retention_days" {
  description = "CloudWatch Logs retention in days for VPC flow logs. Must be one of the values accepted by aws_cloudwatch_log_group (e.g. 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653, or 0 for never expire)."
  type        = number
  default     = 30

  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653, 0
    ], var.flow_log_retention_days)
    error_message = "flow_log_retention_days must be a value supported by aws_cloudwatch_log_group retention_in_days."
  }
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}

