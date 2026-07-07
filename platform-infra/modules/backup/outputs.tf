output "bucket_name" {
  description = "Name of the S3 bucket Velero stores cluster backups in."
  value       = aws_s3_bucket.velero.bucket
}

output "bucket_arn" {
  description = "ARN of the S3 bucket Velero stores cluster backups in."
  value       = aws_s3_bucket.velero.arn
}

output "role_arn" {
  description = "IAM role ARN assumed by Velero's service account via IRSA."
  value       = aws_iam_role.velero.arn
}

output "namespace" {
  description = "Namespace Velero is installed into."
  value       = helm_release.velero.namespace
}
