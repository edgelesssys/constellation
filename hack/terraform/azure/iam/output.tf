output "subscription" {
  value = data.azurerm_subscription.current.subscription_id
}

output "tenant" {
  value = data.azurerm_subscription.current.tenant_id
}

output "location" {
  value = var.region
}

output "resourceGroup" {
  value = var.resource_group_name
}

output "appClientID" {
  value = azuread_application.base_application.application_id
}

output "userAssignedIdentity" {
  value = azurerm_user_assigned_identity.identity_uami.id
}

output "clientSecretValue" {
    value = azuread_application_password.base_application_secret.value
    sensitive = true
}
