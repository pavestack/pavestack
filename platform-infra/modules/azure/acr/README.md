# azure/acr module

Creates an Azure Container Registry (Premium SKU) with admin access disabled, anonymous pull disabled, content trust enabled, and an untagged-manifest retention policy — the Azure counterpart to the AWS `ecr` module's immutable, scan-on-push posture.

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
| azurerm_container_registry.this | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_location"></a> [location](#input\_location) | Azure region. | `string` | n/a | yes |
| <a name="input_registry_name"></a> [registry\_name](#input\_registry\_name) | Azure Container Registry name (alphanumeric, globally unique). | `string` | n/a | yes |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | Resource group that owns the registry. | `string` | n/a | yes |
| <a name="input_retention_days"></a> [retention\_days](#input\_retention\_days) | Number of days to retain untagged manifests. | `number` | `30` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_login_server"></a> [login\_server](#output\_login\_server) | Container registry login server hostname. |
| <a name="output_registry_id"></a> [registry\_id](#output\_registry\_id) | Container registry resource ID. |
| <a name="output_registry_name"></a> [registry\_name](#output\_registry\_name) | Container registry name. |
<!-- END_TF_DOCS -->
