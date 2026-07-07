# ingress module

Deploys the cluster's ingress edge stack: the AWS Load Balancer Controller (with a reproduction of the upstream IAM policy), cert-manager, and external-dns — each independently toggleable and each with its own IRSA role where AWS access is required. `ClusterIssuer` resources are intentionally not created here; see `platform-config/templates/cluster-issuer`.

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
| aws_iam_policy.aws_load_balancer_controller | resource |
| aws_iam_policy.external_dns | resource |
| aws_iam_policy_document.aws_load_balancer_controller_assume_role | data source |
| aws_iam_policy_document.external_dns | data source |
| aws_iam_policy_document.external_dns_assume_role | data source |
| aws_iam_role.aws_load_balancer_controller | resource |
| aws_iam_role.external_dns | resource |
| aws_iam_role_policy_attachment.aws_load_balancer_controller | resource |
| aws_iam_role_policy_attachment.external_dns | resource |
| helm_release.aws_load_balancer_controller | resource |
| helm_release.cert_manager | resource |
| helm_release.external_dns | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_aws_load_balancer_controller_chart_version"></a> [aws\_load\_balancer\_controller\_chart\_version](#input\_aws\_load\_balancer\_controller\_chart\_version) | aws-load-balancer-controller Helm chart version (from https://aws.github.io/eks-charts). | `string` | `"1.8.1"` | no |
| <a name="input_aws_load_balancer_controller_values"></a> [aws\_load\_balancer\_controller\_values](#input\_aws\_load\_balancer\_controller\_values) | Additional Helm values (raw YAML strings) merged over the AWS Load Balancer Controller defaults. | `list(string)` | `[]` | no |
| <a name="input_cert_manager_chart_version"></a> [cert\_manager\_chart\_version](#input\_cert\_manager\_chart\_version) | cert-manager Helm chart version (from https://charts.jetstack.io). | `string` | `"v1.15.3"` | no |
| <a name="input_cert_manager_values"></a> [cert\_manager\_values](#input\_cert\_manager\_values) | Additional Helm values (raw YAML strings) merged over the cert-manager defaults. | `list(string)` | `[]` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | EKS cluster name. Used for IAM role naming, the AWS Load Balancer Controller's clusterName value, and external-dns's txtOwnerId. | `string` | n/a | yes |
| <a name="input_domain_filter"></a> [domain\_filter](#input\_domain\_filter) | Domain suffix external-dns restricts record management to (e.g. apps.example.com). | `string` | n/a | yes |
| <a name="input_enable_aws_load_balancer_controller"></a> [enable\_aws\_load\_balancer\_controller](#input\_enable\_aws\_load\_balancer\_controller) | Deploy the AWS Load Balancer Controller and its IRSA role. | `bool` | `true` | no |
| <a name="input_enable_cert_manager"></a> [enable\_cert\_manager](#input\_enable\_cert\_manager) | Deploy cert-manager. | `bool` | `true` | no |
| <a name="input_enable_external_dns"></a> [enable\_external\_dns](#input\_enable\_external\_dns) | Deploy external-dns and its IRSA role. | `bool` | `true` | no |
| <a name="input_external_dns_chart_version"></a> [external\_dns\_chart\_version](#input\_external\_dns\_chart\_version) | external-dns Helm chart version (from https://kubernetes-sigs.github.io/external-dns). | `string` | `"1.15.0"` | no |
| <a name="input_external_dns_policy"></a> [external\_dns\_policy](#input\_external\_dns\_policy) | external-dns record management policy. upsert-only never deletes records, even if the source resource is removed; sync keeps records in lockstep with cluster state. | `string` | `"upsert-only"` | no |
| <a name="input_external_dns_values"></a> [external\_dns\_values](#input\_external\_dns\_values) | Additional Helm values (raw YAML strings) merged over the external-dns defaults. | `list(string)` | `[]` | no |
| <a name="input_oidc_issuer_url"></a> [oidc\_issuer\_url](#input\_oidc\_issuer\_url) | Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys. | `string` | n/a | yes |
| <a name="input_oidc_provider_arn"></a> [oidc\_provider\_arn](#input\_oidc\_provider\_arn) | ARN of the IAM OIDC identity provider registered for the EKS cluster (aws\_iam\_openid\_connect\_provider). Required for IRSA trust policies. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | AWS region the EKS cluster runs in. | `string` | n/a | yes |
| <a name="input_route53_zone_id"></a> [route53\_zone\_id](#input\_route53\_zone\_id) | Route53 hosted zone ID that external-dns is permitted to manage records in. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all IAM resources. | `map(string)` | `{}` | no |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | VPC ID hosting the EKS cluster, passed to the AWS Load Balancer Controller. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_aws_load_balancer_controller_namespace"></a> [aws\_load\_balancer\_controller\_namespace](#output\_aws\_load\_balancer\_controller\_namespace) | Namespace the AWS Load Balancer Controller is installed into (null if disabled). |
| <a name="output_aws_load_balancer_controller_role_arn"></a> [aws\_load\_balancer\_controller\_role\_arn](#output\_aws\_load\_balancer\_controller\_role\_arn) | IAM role ARN assumed by the AWS Load Balancer Controller's service account (null if disabled). |
| <a name="output_cert_manager_namespace"></a> [cert\_manager\_namespace](#output\_cert\_manager\_namespace) | Namespace cert-manager is installed into (null if disabled). |
| <a name="output_external_dns_namespace"></a> [external\_dns\_namespace](#output\_external\_dns\_namespace) | Namespace external-dns is installed into (null if disabled). |
| <a name="output_external_dns_role_arn"></a> [external\_dns\_role\_arn](#output\_external\_dns\_role\_arn) | IAM role ARN assumed by external-dns's service account (null if disabled). |
<!-- END_TF_DOCS -->
