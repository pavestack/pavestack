output "role_arn" {
  description = "GitHub Actions role ARN."
  value       = aws_iam_role.this.arn
}

output "role_name" {
  description = "GitHub Actions role name."
  value       = aws_iam_role.this.name
}

