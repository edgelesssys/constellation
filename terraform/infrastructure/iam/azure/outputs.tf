output "subscription_id" {
  value       = data.azurerm_subscription.current.subscription_id
  description = "ID of the Azure subscription."
}

output "tenant_id" {
  value       = data.azurerm_subscription.current.tenant_id
  description = "ID of the Azure tenant."
}

output "uami_id" {
  value       = azurerm_user_assigned_identity.identity_uami.id
  description = "Resource ID of the UAMI in the format: /$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME. Not to be confused with the Client ID of the UAMI."
}

output "base_resource_group" {
  value       = azurerm_resource_group.base_resource_group.name
  description = "Name of the resource group."
}
