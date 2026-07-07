# azure/aks module

Provisions an AKS cluster with workload identity and OIDC issuer enabled, Azure RBAC for Kubernetes authorization, Azure CNI networking, and diagnostics shipped to a dedicated Log Analytics workspace — the Azure counterpart to the AWS `eks` module. Optionally grants the kubelet identity `AcrPull` on a container registry.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | >= 4.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) | >= 4.0 |

## Resources

| Name | Type |
|------|------|
| azurerm_kubernetes_cluster.this | resource |
| azurerm_log_analytics_workspace.this | resource |
| azurerm_role_assignment.acr_pull | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_acr_id"></a> [acr\_id](#input\_acr\_id) | Container registry resource ID to grant AcrPull to the kubelet identity. Empty to skip. | `string` | `""` | no |
| <a name="input_admin_group_object_ids"></a> [admin\_group\_object\_ids](#input\_admin\_group\_object\_ids) | Entra ID group object IDs granted cluster admin via Azure RBAC. | `list(string)` | `[]` | no |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version for the AKS control plane. | `string` | n/a | yes |
| <a name="input_location"></a> [location](#input\_location) | Azure region. | `string` | n/a | yes |
| <a name="input_log_retention_days"></a> [log\_retention\_days](#input\_log\_retention\_days) | Log Analytics workspace retention in days. | `number` | `30` | no |
| <a name="input_name"></a> [name](#input\_name) | AKS cluster name. | `string` | n/a | yes |
| <a name="input_node_max_count"></a> [node\_max\_count](#input\_node\_max\_count) | Maximum default node pool size. | `number` | `3` | no |
| <a name="input_node_min_count"></a> [node\_min\_count](#input\_node\_min\_count) | Minimum default node pool size. | `number` | `1` | no |
| <a name="input_node_vm_size"></a> [node\_vm\_size](#input\_node\_vm\_size) | VM size for the default node pool. | `string` | `"Standard_D2s_v5"` | no |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | Resource group that owns the cluster. | `string` | n/a | yes |
| <a name="input_subnet_id"></a> [subnet\_id](#input\_subnet\_id) | Subnet ID for the default node pool. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_client_certificate"></a> [client\_certificate](#output\_client\_certificate) | Base64 encoded client certificate for cluster admin. |
| <a name="output_client_key"></a> [client\_key](#output\_client\_key) | Base64 encoded client key for cluster admin. |
| <a name="output_cluster_ca_certificate"></a> [cluster\_ca\_certificate](#output\_cluster\_ca\_certificate) | Base64 encoded AKS cluster CA certificate. |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | AKS cluster API server host. |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | AKS cluster resource ID. |
| <a name="output_cluster_name"></a> [cluster\_name](#output\_cluster\_name) | AKS cluster name. |
| <a name="output_cluster_oidc_issuer_url"></a> [cluster\_oidc\_issuer\_url](#output\_cluster\_oidc\_issuer\_url) | AKS OIDC issuer URL for workload identity. |
| <a name="output_kubelet_identity_object_id"></a> [kubelet\_identity\_object\_id](#output\_kubelet\_identity\_object\_id) | Object ID of the kubelet managed identity. |
<!-- END_TF_DOCS -->
