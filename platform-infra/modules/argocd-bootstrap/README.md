# argocd-bootstrap module

Installs Argo CD via the upstream Helm chart with metrics and ServiceMonitors enabled so the observability module's Argo CD alert rules have something to scrape. Bootstraps GitOps delivery for the rest of the platform; tenant applications and ApplicationSets live in `platform-config`, not here.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | >= 2.17 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_helm"></a> [helm](#provider\_helm) | >= 2.17 |

## Resources

| Name | Type |
|------|------|
| helm_release.argocd | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | Argo CD Helm chart version. | `string` | n/a | yes |
| <a name="input_namespace"></a> [namespace](#input\_namespace) | Namespace for Argo CD. | `string` | `"argocd"` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values for Argo CD. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Argo CD namespace. |
| <a name="output_release_name"></a> [release\_name](#output\_release\_name) | Argo CD Helm release name. |
<!-- END_TF_DOCS -->
