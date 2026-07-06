output "vnet_id" {
  description = "Virtual network ID."
  value       = azurerm_virtual_network.this.id
}

output "vnet_name" {
  description = "Virtual network name."
  value       = azurerm_virtual_network.this.name
}

output "aks_subnet_id" {
  description = "Subnet ID used by the AKS node pool."
  value       = azurerm_subnet.aks.id
}

output "subnet_ids" {
  description = "All subnet IDs in the virtual network."
  value       = [azurerm_subnet.aks.id]
}
