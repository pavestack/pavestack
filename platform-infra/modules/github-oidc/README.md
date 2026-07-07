# github-oidc module

Creates a GitHub Actions IAM role trusted via OIDC (scoped to a single repository and environment) with least-privilege, resource-scoped Terraform plan/apply permissions for the whole platform, split across two managed policies to stay under the IAM policy size limit.

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
| aws_caller_identity.current | data source |
| aws_iam_openid_connect_provider.github | resource |
| aws_iam_policy.terraform_bootstrap | resource |
| aws_iam_policy.terraform_bootstrap_iam | resource |
| aws_iam_policy_document.terraform_bootstrap | data source |
| aws_iam_policy_document.terraform_bootstrap_iam | data source |
| aws_iam_policy_document.trust | data source |
| aws_iam_role.this | resource |
| aws_iam_role_policy_attachment.bootstrap | resource |
| aws_iam_role_policy_attachment.bootstrap_iam | resource |
| aws_iam_role_policy_attachment.managed | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_create_oidc_provider"></a> [create\_oidc\_provider](#input\_create\_oidc\_provider) | Create the GitHub OIDC provider in this account. | `bool` | `true` | no |
| <a name="input_ecr_repository_prefix"></a> [ecr\_repository\_prefix](#input\_ecr\_repository\_prefix) | ECR repository name prefix the role may manage and push to (arn:aws:ecr:*:<account>:repository/<prefix>/*). | `string` | `"pavestack"` | no |
| <a name="input_github_environment"></a> [github\_environment](#input\_github\_environment) | GitHub environment allowed to assume the role. | `string` | n/a | yes |
| <a name="input_github_repository"></a> [github\_repository](#input\_github\_repository) | GitHub repository in owner/name form. | `string` | n/a | yes |
| <a name="input_lock_table_arns"></a> [lock\_table\_arns](#input\_lock\_table\_arns) | DynamoDB table ARNs used for Terraform state locking. Empty (the default) grants nothing; the backend uses S3 lockfiles (use\_lockfile) by default. | `list(string)` | `[]` | no |
| <a name="input_managed_policy_arns"></a> [managed\_policy\_arns](#input\_managed\_policy\_arns) | Additional managed policies to attach. | `set(string)` | `[]` | no |
| <a name="input_name"></a> [name](#input\_name) | Role name prefix. | `string` | n/a | yes |
| <a name="input_route53_zone_arns"></a> [route53\_zone\_arns](#input\_route53\_zone\_arns) | Route53 hosted zone ARNs the role may change records in. Empty (the default) grants read-only Route53 access; external-dns manages records at runtime with its own role. | `list(string)` | `[]` | no |
| <a name="input_secret_name_prefixes"></a> [secret\_name\_prefixes](#input\_secret\_name\_prefixes) | Secrets Manager secret name prefixes the role may read (External Secrets Operator sources). | `list(string)` | <pre>[<br/>  "pavestack"<br/>]</pre> | no |
| <a name="input_state_bucket_arns"></a> [state\_bucket\_arns](#input\_state\_bucket\_arns) | S3 bucket ARNs the role may use for Terraform remote state. Empty (the default) derives arn:aws:s3:::<name>-tfstate-* to match the bootstrap/remote-state naming. | `list(string)` | `[]` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_role_arn"></a> [role\_arn](#output\_role\_arn) | GitHub Actions role ARN. |
| <a name="output_role_name"></a> [role\_name](#output\_role\_name) | GitHub Actions role name. |
<!-- END_TF_DOCS -->
