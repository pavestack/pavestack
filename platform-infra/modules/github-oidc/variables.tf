variable "name" {
  description = "Role name prefix."
  type        = string
}

variable "github_repository" {
  description = "GitHub repository in owner/name form."
  type        = string
}

variable "github_environment" {
  description = "GitHub environment allowed to assume the role."
  type        = string
}

variable "create_oidc_provider" {
  description = "Create the GitHub OIDC provider in this account."
  type        = bool
  default     = true
}

variable "managed_policy_arns" {
  description = "Additional managed policies to attach."
  type        = set(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}

