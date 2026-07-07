# azure/github-oidc module

Creates a user-assigned managed identity federated with GitHub Actions OIDC (scoped to a single repository and environment) and grants it an arbitrary set of role assignments — the Azure counterpart to the AWS `github-oidc` module.

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
| azurerm_federated_identity_credential.github | resource |
| azurerm_role_assignment.this | resource |
| azurerm_user_assigned_identity.this | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_github_environment"></a> [github\_environment](#input\_github\_environment) | GitHub environment allowed to federate with the identity. | `string` | n/a | yes |
| <a name="input_github_repository"></a> [github\_repository](#input\_github\_repository) | GitHub repository in owner/name form. | `string` | n/a | yes |
| <a name="input_location"></a> [location](#input\_location) | Azure region. | `string` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | Identity name prefix. | `string` | n/a | yes |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | Resource group that owns the identity. | `string` | n/a | yes |
| <a name="input_role_assignments"></a> [role\_assignments](#input\_role\_assignments) | Map of role assignments (role definition name => scope) granted to the identity. | `map(string)` | `{}` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags applied to all resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_client_id"></a> [client\_id](#output\_client\_id) | Client ID of the GitHub Actions user-assigned identity. |
| <a name="output_identity_id"></a> [identity\_id](#output\_identity\_id) | Resource ID of the GitHub Actions user-assigned identity. |
| <a name="output_principal_id"></a> [principal\_id](#output\_principal\_id) | Principal (object) ID of the GitHub Actions user-assigned identity. |
<!-- END_TF_DOCS -->
