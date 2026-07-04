# Premium SKU is required for retention and trust policies, mirroring the
# immutable, scan-on-push posture of the AWS ECR module.
# checkov:skip=CKV_AZURE_139:Public network access is required for the CLI-first demo platform
# checkov:skip=CKV_AZURE_163:Defender vulnerability scanning requires a paid Defender plan
# checkov:skip=CKV_AZURE_165:Geo-replication is not required for the demo platform
# checkov:skip=CKV_AZURE_166:Image quarantine is not required for the demo platform
# checkov:skip=CKV_AZURE_233:Zone redundancy is not required for the demo platform
# checkov:skip=CKV_AZURE_237:Dedicated data endpoints are not required for the demo platform
resource "azurerm_container_registry" "this" {
  name                          = var.registry_name
  resource_group_name           = var.resource_group_name
  location                      = var.location
  sku                           = "Premium"
  admin_enabled                 = false
  public_network_access_enabled = true
  anonymous_pull_enabled        = false
  retention_policy_in_days      = var.retention_days
  trust_policy_enabled          = true

  tags = var.tags
}
