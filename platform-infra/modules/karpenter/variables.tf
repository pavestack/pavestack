variable "cluster_name" {
  description = "EKS cluster name. Used for IAM role/policy naming, resource tag scoping, and the Karpenter settings.clusterName Helm value."
  type        = string
}

variable "cluster_endpoint" {
  description = "EKS cluster API server endpoint URL, passed to Karpenter as settings.clusterEndpoint."
  type        = string
}

variable "region" {
  description = "AWS region the EKS cluster runs in. Used to scope IAM policy resource ARNs."
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the IAM OIDC identity provider registered for the EKS cluster (aws_iam_openid_connect_provider). Required for the Karpenter controller's IRSA trust policy."
  type        = string
}

variable "oidc_issuer_url" {
  description = "Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys."
  type        = string
}

variable "chart_version" {
  description = "Karpenter Helm chart version (from oci://public.ecr.aws/karpenter)."
  type        = string
  default     = "1.0.8"
}

variable "discovery_subnet_ids" {
  description = "Subnet IDs to tag with karpenter.sh/discovery=<cluster_name> so Karpenter's EC2NodeClass subnet selector can discover them. Leave empty to manage discovery tags elsewhere (e.g. directly on the VPC module's subnets)."
  type        = list(string)
  default     = []
}

variable "discovery_security_group_ids" {
  description = "Security group IDs to tag with karpenter.sh/discovery=<cluster_name> so Karpenter's EC2NodeClass security group selector can discover them."
  type        = list(string)
  default     = []
}

variable "values" {
  description = "Additional Helm values (raw YAML strings) merged over the Karpenter chart defaults."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all AWS resources."
  type        = map(string)
  default     = {}
}
