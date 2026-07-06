locals {
  name = "${var.name_prefix}-${var.environment}"

  tags = {
    Project     = "pavestack"
    Repository  = "platform-infra"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

resource "azurerm_resource_group" "this" {
  name     = var.resource_group_name
  location = var.location

  tags = local.tags
}

module "network" {
  source = "../../../modules/azure/network"

  name                = local.name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  vnet_cidr           = var.vnet_cidr
  tags                = local.tags
}

module "acr" {
  source = "../../../modules/azure/acr"

  registry_name       = var.acr_name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  tags                = local.tags
}

module "github_oidc" {
  count = var.enable_github_oidc_role ? 1 : 0

  source = "../../../modules/azure/github-oidc"

  name                = local.name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  github_repository   = var.github_repository
  github_environment  = var.environment

  role_assignments = {
    "Contributor" = azurerm_resource_group.this.id
  }

  tags = local.tags
}

module "aks" {
  source = "../../../modules/azure/aks"

  name                   = local.name
  location               = azurerm_resource_group.this.location
  resource_group_name    = azurerm_resource_group.this.name
  kubernetes_version     = var.kubernetes_version
  subnet_id              = module.network.aks_subnet_id
  node_vm_size           = var.node_vm_size
  node_min_count         = 1
  node_max_count         = 3
  acr_id                 = module.acr.registry_id
  admin_group_object_ids = var.admin_group_object_ids
  tags                   = local.tags
}

module "argocd" {
  source = "../../../modules/argocd-bootstrap"

  chart_version = "9.5.17"

  depends_on = [module.aks]
}
