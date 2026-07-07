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

variable "state_bucket_arns" {
  description = "S3 bucket ARNs the role may use for Terraform remote state. Empty (the default) derives arn:aws:s3:::<name>-tfstate-* to match the bootstrap/remote-state naming."
  type        = list(string)
  default     = []
}

variable "lock_table_arns" {
  description = "DynamoDB table ARNs used for Terraform state locking. Empty (the default) grants nothing; the backend uses S3 lockfiles (use_lockfile) by default."
  type        = list(string)
  default     = []
}

variable "ecr_repository_prefix" {
  description = "ECR repository name prefix the role may manage and push to (arn:aws:ecr:*:<account>:repository/<prefix>/*)."
  type        = string
  default     = "pavestack"
}

variable "secret_name_prefixes" {
  description = "Secrets Manager secret name prefixes the role may read (External Secrets Operator sources)."
  type        = list(string)
  default     = ["pavestack"]
}

variable "route53_zone_arns" {
  description = "Route53 hosted zone ARNs the role may change records in. Empty (the default) grants read-only Route53 access; external-dns manages records at runtime with its own role."
  type        = list(string)
  default     = []
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

