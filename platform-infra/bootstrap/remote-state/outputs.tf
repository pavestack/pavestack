output "state_bucket" {
  description = "S3 bucket for Terraform state."
  value       = aws_s3_bucket.state.id
}


output "kms_key_id" {
  description = "KMS key ID for Terraform state encryption."
  value       = aws_kms_key.state.key_id
}

