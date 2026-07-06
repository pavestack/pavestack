locals {
  tags = merge(var.tags, {
    Project    = "pavestack"
    Repository = "platform-infra"
    ManagedBy  = "terraform"
    Purpose    = "terraform-remote-state"
  })
}

resource "azurerm_resource_group" "state" {
  name     = var.resource_group_name
  location = var.location

  tags = local.tags
}

# The azurerm backend authenticates to this account with a shared key, so shared
# key access and public network access stay enabled. Data is protected with GRS
# replication, blob versioning, soft delete, and TLS 1.2.
# checkov:skip=CKV_AZURE_35:Public network access is required for the remote state backend
# checkov:skip=CKV_AZURE_59:Public network access is required for the remote state backend
# checkov:skip=CKV_AZURE_33:Queue service logging is not applicable to a state-only account
# checkov:skip=CKV2_AZURE_1:Customer-managed key encryption is out of scope for the demo platform
# checkov:skip=CKV2_AZURE_33:Private endpoints are out of scope for the demo platform
# checkov:skip=CKV2_AZURE_40:Shared key access is required by the azurerm backend
# checkov:skip=CKV2_AZURE_41:SAS expiration policy is not applicable to a state-only account
resource "azurerm_storage_account" "state" {
  name                            = var.storage_account_name
  resource_group_name             = azurerm_resource_group.state.name
  location                        = azurerm_resource_group.state.location
  account_tier                    = "Standard"
  account_replication_type        = "GRS"
  account_kind                    = "StorageV2"
  min_tls_version                 = "TLS1_2"
  https_traffic_only_enabled      = true
  allow_nested_items_to_be_public = false
  shared_access_key_enabled       = true
  public_network_access_enabled   = true

  blob_properties {
    versioning_enabled = true

    delete_retention_policy {
      days = var.retention_days
    }

    container_delete_retention_policy {
      days = var.retention_days
    }
  }

  tags = local.tags
}

resource "azurerm_storage_container" "state" {
  name                  = var.container_name
  storage_account_id    = azurerm_storage_account.state.id
  container_access_type = "private"
}
