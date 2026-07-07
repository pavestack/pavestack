# finops module

Deploys OpenCost for in-cluster cost allocation, reading usage metrics from the observability module's Prometheus and attributing spend to namespaces/tenants. Also creates an AWS Budgets monthly cost budget with forecasted and actual overspend alerts published to an SNS topic (with optional email subscriptions).

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
| aws_budgets_budget.monthly | resource |
| aws_caller_identity.current | data source |
| aws_iam_policy_document.budget_alerts_publish | data source |
| aws_sns_topic.budget_alerts | resource |
| aws_sns_topic_policy.budget_alerts | resource |
| aws_sns_topic_subscription.budget_alerts_email | resource |
| helm_release.opencost | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_budget_notification_emails"></a> [budget\_notification\_emails](#input\_budget\_notification\_emails) | Email addresses subscribed to AWS Budgets alerts (forecasted and actual overspend). | `list(string)` | `[]` | no |
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | OpenCost Helm chart version. | `string` | `"1.40.0"` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | EKS cluster name, reported to OpenCost as the default cluster ID for cost allocation. | `string` | n/a | yes |
| <a name="input_monthly_budget_amount"></a> [monthly\_budget\_amount](#input\_monthly\_budget\_amount) | Monthly AWS cost budget limit, in USD. | `string` | `"100"` | no |
| <a name="input_prometheus_namespace"></a> [prometheus\_namespace](#input\_prometheus\_namespace) | Namespace the Prometheus Service referenced by prometheus\_service\_name runs in. | `string` | n/a | yes |
| <a name="input_prometheus_port"></a> [prometheus\_port](#input\_prometheus\_port) | Port the Prometheus Service listens on. | `number` | `9090` | no |
| <a name="input_prometheus_service_name"></a> [prometheus\_service\_name](#input\_prometheus\_service\_name) | Kubernetes Service name of the in-cluster Prometheus that OpenCost queries for usage metrics (see the observability module's kube-prometheus-stack release, e.g. kube-prometheus-stack-prometheus). | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all AWS resources created by this module. | `map(string)` | `{}` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values for the OpenCost release. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_budget_name"></a> [budget\_name](#output\_budget\_name) | Name of the AWS Budgets monthly cost budget. |
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Namespace OpenCost is deployed into. |
| <a name="output_sns_topic_arn"></a> [sns\_topic\_arn](#output\_sns\_topic\_arn) | ARN of the SNS topic AWS Budgets notifications are published to. |
<!-- END_TF_DOCS -->
