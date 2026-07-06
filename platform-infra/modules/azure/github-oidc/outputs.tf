output "client_id" {
  description = "Client ID of the GitHub Actions user-assigned identity."
  value       = azurerm_user_assigned_identity.this.client_id
}

output "principal_id" {
  description = "Principal (object) ID of the GitHub Actions user-assigned identity."
  value       = azurerm_user_assigned_identity.this.principal_id
}

output "identity_id" {
  description = "Resource ID of the GitHub Actions user-assigned identity."
  value       = azurerm_user_assigned_identity.this.id
}
