output "subscription_id" {
  value = data.azurerm_subscription.current.subscription_id
  description = "Azure subscription ID."
}

output "tenant_id" {
  value = data.azurerm_subscription.current.tenant_id
  description = "Azure tenant ID."
}

output "uami_id" {
  description = "Outputs the id in the format: /$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME. Not to be confused with the client_id."
  value       = azurerm_user_assigned_identity.identity_uami.id
}

output "base_resource_group" {
  value = azurerm_resource_group.base_resource_group.name
  description = "The name of the resource group, as provided by the input variable `resource_group_name`."
}
