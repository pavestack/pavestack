resource "azurerm_log_analytics_workspace" "this" {
  name                = "${var.name}-logs"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = var.tags
}

# AKS encrypts etcd secrets at rest with platform-managed keys by default, which
# is the Azure parallel to the EKS module's explicit KMS key. Customer-managed
# key (Key Vault) encryption can be layered on later if required.
# Local accounts stay enabled so Terraform can bootstrap Argo CD with the cluster
# admin kubeconfig, mirroring the EKS module's admin-credential bootstrap.
# checkov:skip=CKV_AZURE_6:API server authorized IP ranges require known egress CIDRs not available in the demo platform
# checkov:skip=CKV_AZURE_115:A private cluster is not used so the CLI-first demo stays reachable
# checkov:skip=CKV_AZURE_117:Customer-managed disk encryption set is not used for the demo platform
# checkov:skip=CKV_AZURE_141:Local accounts stay enabled for the Terraform Argo CD bootstrap
# checkov:skip=CKV_AZURE_170:Paid (Standard) tier SLA is not required for the demo platform
# checkov:skip=CKV_AZURE_226:Ephemeral OS disks are not required for the demo platform
# checkov:skip=CKV_AZURE_227:Host-based encryption is not required for the demo platform
# checkov:skip=CKV_AZURE_232:Restricting the system node pool to critical add-ons is out of scope for the demo platform
resource "azurerm_kubernetes_cluster" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = var.name
  kubernetes_version  = var.kubernetes_version

  oidc_issuer_enabled               = true
  workload_identity_enabled         = true
  role_based_access_control_enabled = true
  azure_policy_enabled              = true

  default_node_pool {
    name                 = "default"
    vm_size              = var.node_vm_size
    vnet_subnet_id       = var.subnet_id
    auto_scaling_enabled = true
    min_count            = var.node_min_count
    max_count            = var.node_max_count
    orchestrator_version = var.kubernetes_version

    node_labels = {
      "pavestack.io/node-pool" = "default"
    }

    upgrade_settings {
      max_surge = "10%"
    }
  }

  identity {
    type = "SystemAssigned"
  }

  azure_active_directory_role_based_access_control {
    admin_group_object_ids = var.admin_group_object_ids
    azure_rbac_enabled     = true
  }

  network_profile {
    network_plugin    = "azure"
    network_policy    = "azure"
    load_balancer_sku = "standard"
  }

  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
  }

  tags = var.tags
}

resource "azurerm_role_assignment" "acr_pull" {
  count = var.acr_id == "" ? 0 : 1

  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = azurerm_kubernetes_cluster.this.kubelet_identity[0].object_id
  skip_service_principal_aad_check = true
}
