output "aws_load_balancer_controller_role_arn" {
  description = "IAM role ARN assumed by the AWS Load Balancer Controller's service account (null if disabled)."
  value       = var.enable_aws_load_balancer_controller ? aws_iam_role.aws_load_balancer_controller[0].arn : null
}

output "aws_load_balancer_controller_namespace" {
  description = "Namespace the AWS Load Balancer Controller is installed into (null if disabled)."
  value       = var.enable_aws_load_balancer_controller ? helm_release.aws_load_balancer_controller[0].namespace : null
}

output "cert_manager_namespace" {
  description = "Namespace cert-manager is installed into (null if disabled)."
  value       = var.enable_cert_manager ? helm_release.cert_manager[0].namespace : null
}

output "external_dns_role_arn" {
  description = "IAM role ARN assumed by external-dns's service account (null if disabled)."
  value       = var.enable_external_dns ? aws_iam_role.external_dns[0].arn : null
}

output "external_dns_namespace" {
  description = "Namespace external-dns is installed into (null if disabled)."
  value       = var.enable_external_dns ? helm_release.external_dns[0].namespace : null
}
