# secrets module

Installs the External Secrets Operator (ESO) with an IRSA role granting least-privilege read access to the `<secret_path_prefix>/<tenant>/<name>` Secrets Manager naming convention. `ClusterSecretStore` and `ExternalSecret` resources are intentionally not created here; see `platform-config/templates/{cluster-secret-store,external-secret}`.

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
| aws_iam_policy.external_secrets | resource |
| aws_iam_policy_document.external_secrets | data source |
| aws_iam_policy_document.external_secrets_assume_role | data source |
| aws_iam_role.external_secrets | resource |
| aws_iam_role_policy_attachment.external_secrets | resource |
| helm_release.external_secrets | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | external-secrets Helm chart version (from https://charts.external-secrets.io). | `string` | `"0.10.4"` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | EKS cluster name. Used for IAM role naming. | `string` | n/a | yes |
| <a name="input_oidc_issuer_url"></a> [oidc\_issuer\_url](#input\_oidc\_issuer\_url) | Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys. | `string` | n/a | yes |
| <a name="input_oidc_provider_arn"></a> [oidc\_provider\_arn](#input\_oidc\_provider\_arn) | ARN of the IAM OIDC identity provider registered for the EKS cluster (aws\_iam\_openid\_connect\_provider). Required for IRSA trust policies. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | AWS region the EKS cluster runs in. Scopes the IAM policy's Secrets Manager ARN and configures the ClusterSecretStore's AWS provider region (set separately in platform-config). | `string` | n/a | yes |
| <a name="input_secret_path_prefix"></a> [secret\_path\_prefix](#input\_secret\_path\_prefix) | Secrets Manager path prefix the ESO controller's IAM role is permitted to read. Secrets are expected to follow the <secret\_path\_prefix>/<tenant>/<name> naming convention. | `string` | `"pavestack"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all IAM resources. | `map(string)` | `{}` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values (raw YAML strings) merged over the external-secrets defaults. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_external_secrets_namespace"></a> [external\_secrets\_namespace](#output\_external\_secrets\_namespace) | Namespace the External Secrets Operator is installed into. |
| <a name="output_external_secrets_role_arn"></a> [external\_secrets\_role\_arn](#output\_external\_secrets\_role\_arn) | IAM role ARN assumed by the External Secrets Operator controller's service account. |
<!-- END_TF_DOCS -->
