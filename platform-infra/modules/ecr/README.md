# ecr module

Creates a set of ECR repositories with immutable image tags, KMS encryption, and scan-on-push enabled, plus a lifecycle policy that keeps only the most recent 50 images per repository.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 6.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 6.0 |

## Resources

| Name | Type |
|------|------|
| aws_ecr_lifecycle_policy.this | resource |
| aws_ecr_repository.this | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_repositories"></a> [repositories](#input\_repositories) | ECR repository names to create. | `set(string)` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_repository_arns"></a> [repository\_arns](#output\_repository\_arns) | Repository ARNs keyed by repository name. |
| <a name="output_repository_urls"></a> [repository\_urls](#output\_repository\_urls) | Repository URLs keyed by repository name. |
<!-- END_TF_DOCS -->
