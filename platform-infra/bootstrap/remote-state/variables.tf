variable "name_prefix" {
  description = "Prefix for remote state resources."
  type        = string
}

variable "environment" {
  description = "Bootstrap environment label."
  type        = string
  default     = "shared"
}

variable "aws_region" {
  description = "AWS region for state resources."
  type        = string
  default     = "eu-central-1"
}

variable "force_destroy" {
  description = "Allow deleting a non-empty state bucket. Keep false outside experiments."
  type        = bool
  default     = false
}

variable "cost_center" {
  description = "FinOps cost-center tag applied to every resource this environment provisions. See AGENTS.md 'cost-tagging convention'."
  type        = string
  default     = "platform-engineering"
}

variable "team" {
  description = "Owning team tag, matching the pavestack.io/team label convention used in platform-config."
  type        = string
  default     = "platform"
}

