variable "cluster_name" {
  description = "EKS cluster name. Used for the Velero backup bucket name, KMS alias, IAM role naming, and the Velero backup storage location prefix."
  type        = string
}

variable "region" {
  description = "AWS region the EKS cluster runs in."
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
  description = "Velero Helm chart version (from https://vmware-tanzu.github.io/helm-charts)."
  type        = string
  default     = "7.2.1"
}

variable "backup_retention_days" {
  description = "Number of days after which objects in the Velero backup bucket expire via S3 lifecycle rule."
  type        = number
  default     = 90
}

variable "backup_schedule" {
  description = "Cron expression for Velero's default daily-cluster backup schedule."
  type        = string
  default     = "0 3 * * *"
}

variable "enable_node_agent" {
  description = "Deploy the Velero node-agent DaemonSet (fs-backup / restic/kopia path). Leave disabled unless file-system-level backups of non-EBS volumes are required."
  type        = bool
  default     = false
}

variable "values" {
  description = "Additional Helm values (raw YAML strings) merged over the Velero defaults."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all AWS resources."
  type        = map(string)
  default     = {}
}
