# policy module

Installs the Kyverno admission controller, which validates the baseline `ClusterPolicy` set delivered via GitOps at `platform-config/policies`. Kyverno only talks to the Kubernetes API server it runs inside, so no AWS IAM is required.

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
| helm_release.kyverno | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_chart_version"></a> [chart\_version](#input\_chart\_version) | Kyverno Helm chart version. | `string` | n/a | yes |
| <a name="input_namespace"></a> [namespace](#input\_namespace) | Namespace for Kyverno. | `string` | `"kyverno"` | no |
| <a name="input_values"></a> [values](#input\_values) | Additional Helm values for Kyverno (e.g. override admissionController.replicas for HA in prod). | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Kyverno namespace. |
| <a name="output_release_name"></a> [release\_name](#output\_release\_name) | Kyverno Helm release name. |
<!-- END_TF_DOCS -->
