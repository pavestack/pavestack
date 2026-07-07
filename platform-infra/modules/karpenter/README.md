# karpenter module

Deploys Karpenter for just-in-time EC2 node autoscaling: an IRSA controller role with the official Karpenter v1 IAM policy, an IAM role for nodes Karpenter launches, an SQS interruption queue wired to EventBridge rules for spot interruption/rebalance/instance-state-change/health events, and optional discovery tagging of subnets and security groups.

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
| aws_cloudwatch_event_rule.instance_state_change | resource |
| aws_cloudwatch_event_rule.rebalance_recommendation | resource |
| aws_cloudwatch_event_rule.scheduled_change | resource |
| aws_cloudwatch_event_rule.spot_interruption | resource |
| aws_cloudwatch_event_target.instance_state_change | resource |
| aws_cloudwatch_event_target.rebalance_recommendation | resource |
| aws_cloudwatch_event_target.scheduled_change | resource |
| aws_cloudwatch_event_target.spot_interruption | resource |
| aws_ec2_tag.discovery_security_group | resource |
| aws_ec2_tag.discovery_subnet | resource |
| aws_eks_access_entry.node | resource |
| aws_iam_policy.controller | resource |
| aws_iam_policy_document.controller_assume_role | data source |
| aws_iam_policy_document.interruption_queue | data source |
| aws_iam_policy_document.node_assume_role | data source |
| aws_iam_role.controller | resource |
| aws_iam_role.node | resource |
| aws_iam_role_policy_attachment.controller | resource |
| aws_iam_role_policy_attachment.node | resource |
| aws_sqs_queue.interruption | resource |
| aws_sqs_queue_policy.interruption | resource |
| helm_release.karpenter | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | Karpenter Helm chart version (from oci://public.ecr.aws/karpenter). | `string` | `"1.0.8"` | no |
| <a name="input_cluster_endpoint"></a> [cluster\_endpoint](#input\_cluster\_endpoint) | EKS cluster API server endpoint URL, passed to Karpenter as settings.clusterEndpoint. | `string` | n/a | yes |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | EKS cluster name. Used for IAM role/policy naming, resource tag scoping, and the Karpenter settings.clusterName Helm value. | `string` | n/a | yes |
| <a name="input_discovery_security_group_ids"></a> [discovery\_security\_group\_ids](#input\_discovery\_security\_group\_ids) | Security group IDs to tag with karpenter.sh/discovery=<cluster\_name> so Karpenter's EC2NodeClass security group selector can discover them. | `list(string)` | `[]` | no |
| <a name="input_discovery_subnet_ids"></a> [discovery\_subnet\_ids](#input\_discovery\_subnet\_ids) | Subnet IDs to tag with karpenter.sh/discovery=<cluster\_name> so Karpenter's EC2NodeClass subnet selector can discover them. Leave empty to manage discovery tags elsewhere (e.g. directly on the VPC module's subnets). | `list(string)` | `[]` | no |
| <a name="input_oidc_issuer_url"></a> [oidc\_issuer\_url](#input\_oidc\_issuer\_url) | Full HTTPS issuer URL of the EKS cluster's OIDC provider (e.g. https://oidc.eks.<region>.amazonaws.com/id/<id>). The scheme is stripped internally for IAM trust policy condition keys. | `string` | n/a | yes |
| <a name="input_oidc_provider_arn"></a> [oidc\_provider\_arn](#input\_oidc\_provider\_arn) | ARN of the IAM OIDC identity provider registered for the EKS cluster (aws\_iam\_openid\_connect\_provider). Required for the Karpenter controller's IRSA trust policy. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | AWS region the EKS cluster runs in. Used to scope IAM policy resource ARNs. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all AWS resources. | `map(string)` | `{}` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values (raw YAML strings) merged over the Karpenter chart defaults. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_controller_role_arn"></a> [controller\_role\_arn](#output\_controller\_role\_arn) | ARN of the IAM role assumed by the Karpenter controller's service account (IRSA). |
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Namespace the Karpenter controller is installed into. |
| <a name="output_node_role_arn"></a> [node\_role\_arn](#output\_node\_role\_arn) | ARN of the IAM role attached to nodes launched by Karpenter. |
| <a name="output_node_role_name"></a> [node\_role\_name](#output\_node\_role\_name) | Name of the IAM role attached to nodes launched by Karpenter. |
| <a name="output_queue_name"></a> [queue\_name](#output\_queue\_name) | Name of the SQS queue Karpenter consumes spot interruption / rebalance / instance state-change / health events from. |
<!-- END_TF_DOCS -->
