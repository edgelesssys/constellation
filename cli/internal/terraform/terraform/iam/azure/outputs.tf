output "subscription_id" {
  value = data.azurerm_subscription.current.subscription_id
}

output "tenant_id" {
  value = data.azurerm_subscription.current.tenant_id
}

output "application_id" {
  value = azuread_application.base_application.application_id
}

output "uami_id" {
  value = azurerm_user_assigned_identity.identity_uami.id
}

output "application_client_secret_value" {
  value     = azuread_application_password.base_application_secret.value
  sensitive = true
}
