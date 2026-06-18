variable "repositories" {
  description = "ECR repository names to create."
  type        = set(string)
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}

