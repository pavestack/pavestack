# azure/network module

Creates a virtual network with a dedicated AKS subnet and, by default, a NAT gateway for outbound egress — the Azure counterpart to the AWS `vpc` module.

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
| azurerm_nat_gateway.this | resource |
| azurerm_nat_gateway_public_ip_association.this | resource |
| azurerm_public_ip.nat | resource |
| azurerm_subnet.aks | resource |
| azurerm_subnet_nat_gateway_association.aks | resource |
| azurerm_virtual_network.this | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_enable_nat_gateway"></a> [enable\_nat\_gateway](#input\_enable\_nat\_gateway) | Create a NAT gateway for outbound egress from the AKS subnet. | `bool` | `true` | no |
| <a name="input_location"></a> [location](#input\_location) | Azure region. | `string` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | Name prefix for network resources. | `string` | n/a | yes |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | Resource group that owns the network resources. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |
| <a name="input_vnet_cidr"></a> [vnet\_cidr](#input\_vnet\_cidr) | CIDR block for the virtual network. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_aks_subnet_id"></a> [aks\_subnet\_id](#output\_aks\_subnet\_id) | Subnet ID used by the AKS node pool. |
| <a name="output_subnet_ids"></a> [subnet\_ids](#output\_subnet\_ids) | All subnet IDs in the virtual network. |
| <a name="output_vnet_id"></a> [vnet\_id](#output\_vnet\_id) | Virtual network ID. |
| <a name="output_vnet_name"></a> [vnet\_name](#output\_vnet\_name) | Virtual network name. |
<!-- END_TF_DOCS -->
