output "external_secrets_role_arn" {
  description = "IAM role ARN assumed by the External Secrets Operator controller's service account."
  value       = aws_iam_role.external_secrets.arn
}

output "external_secrets_namespace" {
  description = "Namespace the External Secrets Operator is installed into."
  value       = helm_release.external_secrets.namespace
}
