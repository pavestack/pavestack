output "registry_id" {
  description = "Container registry resource ID."
  value       = azurerm_container_registry.this.id
}

output "login_server" {
  description = "Container registry login server hostname."
  value       = azurerm_container_registry.this.login_server
}

output "registry_name" {
  description = "Container registry name."
  value       = azurerm_container_registry.this.name
}
