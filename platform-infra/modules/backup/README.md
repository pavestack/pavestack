# backup module

Deploys Velero for Kubernetes cluster backup and disaster recovery, backed by a dedicated KMS-encrypted, versioned S3 bucket with a lifecycle expiration rule. Grants Velero's service account least-privilege EBS snapshot and S3/KMS access via IRSA and schedules a daily cluster backup.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 6.0 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | >= 2.17 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 6.0 |
| <a name="provider_helm"></a> [helm](#provider\_helm) | >= 2.17 |

## Resources

| Name | Type |
|------|------|
| aws_caller_identity.current | data source |
| aws_iam_policy.velero | resource |
| aws_iam_policy_document.velero | data source |
| aws_iam_policy_document.velero_assume_role | data source |
| aws_iam_policy_document.velero_bucket | data source |
| aws_iam_role.velero | resource |
| aws_iam_role_policy_attachment.velero | resource |
| aws_kms_alias.velero | resource |
| aws_kms_key.velero | resource |
| aws_s3_bucket.velero | resource |
| aws_s3_bucket_lifecycle_configuration.velero | resource |
| aws_s3_bucket_policy.velero | resource |
| aws_s3_bucket_public_access_block.velero | resource |
| aws_s3_bucket_server_side_encryption_configuration.velero | resource |
| aws_s3_bucket_versioning.velero | resource |
| helm_release.velero | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_backup_retention_days"></a> [backup\_retention\_days](#input\_backup\_retention\_days) | Number of days after which objects in the Velero backup bucket expire via S3 lifecycle rule. | `number` | `90` | no |
| <a name="input_backup_schedule"></a> [backup\_schedule](#input\_backup\_schedule) | Cron expression for Velero's default daily-cluster backup schedule. | `string` | `"0 3 * * *"` | no |
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | Velero Helm chart version (from https://vmware-tanzu.github.io/helm-charts). | `string` | `"7.2.1"` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | EKS cluster name. Used for the Velero backup bucket name, KMS alias, IAM role naming, and the Velero backup storage location prefix. | `string` | n/a | yes |
| <a name="input_enable_node_agent"></a> [enable\_node\_agent](#input\_enable\_node\_agent) | Deploy the Velero node-agent DaemonSet (fs-backup / restic/kopia path). Leave disabled unless file-system-level backups of non-EBS volumes are required. | `bool` | `false` | no |
| <a name="input_oidc_issuer_url"></a> [oidc\_issuer\_url](#input\_oidc\_issuer\_url) | Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys. | `string` | n/a | yes |
| <a name="input_oidc_provider_arn"></a> [oidc\_provider\_arn](#input\_oidc\_provider\_arn) | ARN of the IAM OIDC identity provider registered for the EKS cluster (aws\_iam\_openid\_connect\_provider). Required for IRSA trust policies. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | AWS region the EKS cluster runs in. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all AWS resources. | `map(string)` | `{}` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values (raw YAML strings) merged over the Velero defaults. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_bucket_arn"></a> [bucket\_arn](#output\_bucket\_arn) | ARN of the S3 bucket Velero stores cluster backups in. |
| <a name="output_bucket_name"></a> [bucket\_name](#output\_bucket\_name) | Name of the S3 bucket Velero stores cluster backups in. |
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Namespace Velero is installed into. |
| <a name="output_role_arn"></a> [role\_arn](#output\_role\_arn) | IAM role ARN assumed by Velero's service account via IRSA. |
<!-- END_TF_DOCS -->
