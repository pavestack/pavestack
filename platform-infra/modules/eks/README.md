# eks module

Provisions an EKS cluster with KMS-encrypted secrets, control-plane logging to CloudWatch, a managed node group, the core EKS add-ons (VPC CNI, kube-proxy, CoreDNS, Pod Identity Agent, EBS CSI driver with its own IRSA role), an IAM OIDC provider for IRSA, and optional cluster-admin access entries for platform administrators.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 6.0 |
| <a name="requirement_tls"></a> [tls](#requirement\_tls) | >= 4.3 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 6.0 |
| <a name="provider_tls"></a> [tls](#provider\_tls) | >= 4.3 |

## Resources

| Name | Type |
|------|------|
| aws_caller_identity.current | data source |
| aws_cloudwatch_log_group.cluster | resource |
| aws_eks_access_entry.platform_admins | resource |
| aws_eks_access_policy_association.platform_admins | resource |
| aws_eks_addon.this | resource |
| aws_eks_cluster.this | resource |
| aws_eks_node_group.default | resource |
| aws_iam_openid_connect_provider.cluster | resource |
| aws_iam_policy_document.cluster_assume_role | data source |
| aws_iam_policy_document.ebs_csi_assume_role | data source |
| aws_iam_policy_document.node_assume_role | data source |
| aws_iam_role.cluster | resource |
| aws_iam_role.ebs_csi | resource |
| aws_iam_role.node | resource |
| aws_iam_role_policy_attachment.cluster | resource |
| aws_iam_role_policy_attachment.ebs_csi | resource |
| aws_iam_role_policy_attachment.node | resource |
| aws_kms_alias.cluster | resource |
| aws_kms_key.cluster | resource |
| aws_security_group.cluster | resource |
| tls_certificate.cluster | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_endpoint_private_access"></a> [endpoint\_private\_access](#input\_endpoint\_private\_access) | Enable private EKS API endpoint access. | `bool` | `true` | no |
| <a name="input_endpoint_public_access"></a> [endpoint\_public\_access](#input\_endpoint\_public\_access) | Enable public EKS API endpoint access. | `bool` | `true` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for the EKS control plane. | `string` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | EKS cluster name. | `string` | n/a | yes |
| <a name="input_node_desired_size"></a> [node\_desired\_size](#input\_node\_desired\_size) | Desired managed node group size. | `number` | n/a | yes |
| <a name="input_node_instance_types"></a> [node\_instance\_types](#input\_node\_instance\_types) | Managed node group instance types. | `list(string)` | n/a | yes |
| <a name="input_node_max_size"></a> [node\_max\_size](#input\_node\_max\_size) | Maximum managed node group size. | `number` | n/a | yes |
| <a name="input_node_min_size"></a> [node\_min\_size](#input\_node\_min\_size) | Minimum managed node group size. | `number` | n/a | yes |
| <a name="input_platform_admin_role_arns"></a> [platform\_admin\_role\_arns](#input\_platform\_admin\_role\_arns) | IAM role ARNs granted cluster admin access. | `set(string)` | `[]` | no |
| <a name="input_private_subnet_ids"></a> [private\_subnet\_ids](#input\_private\_subnet\_ids) | Private subnet IDs for EKS. | `list(string)` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | VPC ID. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cluster_ca_certificate"></a> [cluster\_ca\_certificate](#output\_cluster\_ca\_certificate) | Base64 encoded EKS cluster CA certificate. |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | EKS cluster endpoint. |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | EKS cluster name. |
| <a name="output_cluster_oidc_issuer_url"></a> [cluster\_oidc\_issuer\_url](#output\_cluster\_oidc\_issuer\_url) | EKS OIDC issuer URL. |
| <a name="output_cluster_security_group_id"></a> [cluster\_security\_group\_id](#output\_cluster\_security\_group\_id) | EKS cluster security group ID. |
| <a name="output_kms_key_arn"></a> [kms\_key\_arn](#output\_kms\_key\_arn) | EKS secrets encryption KMS key ARN. |
| <a name="output_node_role_arn"></a> [node\_role\_arn](#output\_node\_role\_arn) | Managed node group IAM role ARN. |
| <a name="output_oidc_provider_arn"></a> [oidc\_provider\_arn](#output\_oidc\_provider\_arn) | ARN of the IAM OIDC identity provider registered for the EKS cluster. Used by IRSA trust policies (e.g. ingress module). |
<!-- END_TF_DOCS -->
