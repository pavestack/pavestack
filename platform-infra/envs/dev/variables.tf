variable "aws_region" {
  description = "AWS region."
  type        = string
  default     = "eu-central-1"
}

variable "environment" {
  description = "Environment name."
  type        = string
  default     = "dev"
}

variable "name_prefix" {
  description = "Resource name prefix."
  type        = string
  default     = "pavestack"
}

variable "vpc_cidr" {
  description = "VPC CIDR."
  type        = string
  default     = "10.20.0.0/16"
}

variable "kubernetes_version" {
  description = "EKS Kubernetes version."
  type        = string
  default     = "1.36"
}

variable "github_repository" {
  description = "GitHub repository in owner/name form for OIDC trust."
  type        = string
  default     = "pavestack/platform-infra"
}

variable "enable_github_oidc_role" {
  description = "Create GitHub Actions OIDC role for this environment."
  type        = bool
  default     = false
}

variable "create_github_oidc_provider" {
  description = "Create the GitHub OIDC provider in this account."
  type        = bool
  default     = true
}

variable "platform_admin_role_arns" {
  description = "Additional IAM roles granted EKS admin access."
  type        = set(string)
  default     = []
}

variable "image_repositories" {
  description = "Shared ECR repositories owned by platform-infra."
  type        = set(string)
  default = [
    "pavestack/service-template",
    "pavestack/platform-tools"
  ]
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
