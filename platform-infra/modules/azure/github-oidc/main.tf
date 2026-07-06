locals {
  oidc_issuer = "https://token.actions.githubusercontent.com"
}

resource "azurerm_user_assigned_identity" "this" {
  name                = "${var.name}-github-actions"
  location            = var.location
  resource_group_name = var.resource_group_name

  tags = var.tags
}

resource "azurerm_federated_identity_credential" "github" {
  name                = "${var.name}-github-actions"
  resource_group_name = var.resource_group_name
  parent_id           = azurerm_user_assigned_identity.this.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = local.oidc_issuer
  subject             = "repo:${var.github_repository}:environment:${var.github_environment}"
}

resource "azurerm_role_assignment" "this" {
  for_each = var.role_assignments

  scope                = each.value
  role_definition_name = each.key
  principal_id         = azurerm_user_assigned_identity.this.principal_id
}
