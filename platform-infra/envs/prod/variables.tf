variable "aws_region" {
  description = "AWS region."
  type        = string
  default     = "eu-central-1"
}

variable "environment" {
  description = "Environment name."
  type        = string
  default     = "prod"
}

variable "name_prefix" {
  description = "Resource name prefix."
  type        = string
  default     = "pavestack"
}

variable "vpc_cidr" {
  description = "VPC CIDR."
  type        = string
  default     = "10.30.0.0/16"
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

variable "route53_zone_id" {
  description = "Route53 hosted zone ID external-dns is allowed to manage records in. Empty (the default) disables external-dns so `terraform plan` works without a real zone; real deployments set this via tfvars."
  type        = string
  default     = ""
}

variable "platform_domain" {
  description = "Domain suffix external-dns restricts record management to (external-dns domainFilter)."
  type        = string
  default     = "pavestack.example.com"
}

variable "alert_webhook_url" {
  description = "Slack-compatible incoming webhook URL for Alertmanager notifications. Empty routes alerts to a blackhole receiver so the stack works with no external dependency."
  type        = string
  default     = ""
  sensitive   = true
}
