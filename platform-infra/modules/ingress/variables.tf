variable "cluster_name" {
  description = "EKS cluster name. Used for IAM role naming, the AWS Load Balancer Controller's clusterName value, and external-dns's txtOwnerId."
  type        = string
}

variable "region" {
  description = "AWS region the EKS cluster runs in."
  type        = string
}

variable "vpc_id" {
  description = "VPC ID hosting the EKS cluster, passed to the AWS Load Balancer Controller."
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

variable "route53_zone_id" {
  description = "Route53 hosted zone ID that external-dns is permitted to manage records in."
  type        = string
}

variable "domain_filter" {
  description = "Domain suffix external-dns restricts record management to (e.g. apps.example.com)."
  type        = string
}

variable "enable_aws_load_balancer_controller" {
  description = "Deploy the AWS Load Balancer Controller and its IRSA role."
  type        = bool
  default     = true
}

variable "enable_cert_manager" {
  description = "Deploy cert-manager."
  type        = bool
  default     = true
}

variable "enable_external_dns" {
  description = "Deploy external-dns and its IRSA role."
  type        = bool
  default     = true
}

variable "aws_load_balancer_controller_chart_version" {
  description = "aws-load-balancer-controller Helm chart version (from https://aws.github.io/eks-charts)."
  type        = string
  default     = "1.8.1"
}

variable "cert_manager_chart_version" {
  description = "cert-manager Helm chart version (from https://charts.jetstack.io)."
  type        = string
  default     = "v1.15.3"
}

variable "external_dns_chart_version" {
  description = "external-dns Helm chart version (from https://kubernetes-sigs.github.io/external-dns)."
  type        = string
  default     = "1.15.0"
}

variable "external_dns_policy" {
  description = "external-dns record management policy. upsert-only never deletes records, even if the source resource is removed; sync keeps records in lockstep with cluster state."
  type        = string
  default     = "upsert-only"

  validation {
    condition     = contains(["sync", "upsert-only"], var.external_dns_policy)
    error_message = "external_dns_policy must be one of: sync, upsert-only."
  }
}

variable "aws_load_balancer_controller_values" {
  description = "Additional Helm values (raw YAML strings) merged over the AWS Load Balancer Controller defaults."
  type        = list(string)
  default     = []
}

variable "cert_manager_values" {
  description = "Additional Helm values (raw YAML strings) merged over the cert-manager defaults."
  type        = list(string)
  default     = []
}

variable "external_dns_values" {
  description = "Additional Helm values (raw YAML strings) merged over the external-dns defaults."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all IAM resources."
  type        = map(string)
  default     = {}
}
