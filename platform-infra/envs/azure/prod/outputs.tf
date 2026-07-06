output "resource_group_name" {
  description = "Resource group name."
  value       = azurerm_resource_group.this.name
}

output "cluster_name" {
  description = "AKS cluster name."
  value       = module.aks.cluster_name
}

output "cluster_endpoint" {
  description = "AKS cluster endpoint."
  value       = module.aks.cluster_endpoint
}

output "cluster_oidc_issuer_url" {
  description = "AKS OIDC issuer URL."
  value       = module.aks.cluster_oidc_issuer_url
}

output "vnet_id" {
  description = "Virtual network ID."
  value       = module.network.vnet_id
}

output "aks_subnet_id" {
  description = "AKS subnet ID."
  value       = module.network.aks_subnet_id
}

output "acr_login_server" {
  description = "Container registry login server."
  value       = module.acr.login_server
}

output "argocd_namespace" {
  description = "Argo CD namespace."
  value       = module.argocd.namespace
}

output "github_actions_client_id" {
  description = "GitHub Actions federated identity client ID, if created."
  value       = var.enable_github_oidc_role ? module.github_oidc[0].client_id : null
}
