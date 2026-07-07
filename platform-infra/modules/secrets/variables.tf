variable "cluster_name" {
  description = "EKS cluster name. Used for IAM role naming."
  type        = string
}

variable "region" {
  description = "AWS region the EKS cluster runs in. Scopes the IAM policy's Secrets Manager ARN and configures the ClusterSecretStore's AWS provider region (set separately in platform-config)."
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the IAM OIDC identity provider registered for the EKS cluster (aws_iam_openid_connect_provider). Required for IRSA trust policies."
  type        = string
}

variable "oidc_issuer_url" {
  description = "Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys."
  type        = string
}

variable "chart_version" {
  description = "external-secrets Helm chart version (from https://charts.external-secrets.io)."
  type        = string
  default     = "0.10.4"
}

variable "secret_path_prefix" {
  description = "Secrets Manager path prefix the ESO controller's IAM role is permitted to read. Secrets are expected to follow the <secret_path_prefix>/<tenant>/<name> naming convention."
  type        = string
  default     = "pavestack"
}

variable "values" {
  description = "Additional Helm values (raw YAML strings) merged over the external-secrets defaults."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all IAM resources."
  type        = map(string)
  default     = {}
}
