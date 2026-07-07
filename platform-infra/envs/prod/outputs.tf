output "cluster_name" {
  description = "EKS cluster name."
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint."
  value       = module.eks.cluster_endpoint
}

output "endpoint" {
  description = "Alias for cluster_endpoint."
  value       = module.eks.cluster_endpoint
}

output "cluster_oidc_issuer_url" {
  description = "EKS OIDC issuer URL."
  value       = module.eks.cluster_oidc_issuer_url
}

output "vpc_id" {
  description = "VPC ID."
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs."
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "Public subnet IDs."
  value       = module.vpc.public_subnet_ids
}

output "subnet_ids" {
  description = "All subnet IDs (public and private)."
  value       = concat(module.vpc.public_subnet_ids, module.vpc.private_subnet_ids)
}

output "subnets" {
  description = "Alias for subnet_ids."
  value       = concat(module.vpc.public_subnet_ids, module.vpc.private_subnet_ids)
}

output "ecr_repository_urls" {
  description = "ECR repository URLs keyed by name."
  value       = module.ecr.repository_urls
}

output "ecr_urls" {
  description = "Alias for ecr_repository_urls."
  value       = module.ecr.repository_urls
}

output "argocd_namespace" {
  description = "Argo CD namespace."
  value       = module.argocd.namespace
}

output "github_actions_role_arn" {
  description = "GitHub Actions role ARN, if created."
  value       = var.enable_github_oidc_role ? module.github_oidc[0].role_arn : null
}

output "observability_namespace" {
  description = "Namespace the observability stack is deployed into."
  value       = module.observability.namespace
}

output "otel_exporter_otlp_endpoint" {
  description = "Tempo OTLP/HTTP endpoint for services' OTEL_EXPORTER_OTLP_ENDPOINT."
  value       = module.observability.otel_exporter_otlp_endpoint
}

output "grafana_service" {
  description = "Grafana Kubernetes Service name."
  value       = module.observability.grafana_service
}

output "aws_load_balancer_controller_role_arn" {
  description = "IAM role ARN assumed by the AWS Load Balancer Controller."
  value       = module.ingress.aws_load_balancer_controller_role_arn
}

output "external_secrets_role_arn" {
  description = "IAM role ARN assumed by the External Secrets Operator controller's service account."
  value       = module.secrets.external_secrets_role_arn
}

output "external_secrets_namespace" {
  description = "Namespace the External Secrets Operator is installed into."
  value       = module.secrets.external_secrets_namespace
}

output "kyverno_namespace" {
  description = "Namespace the Kyverno admission controller is installed into."
  value       = module.policy.namespace
}
