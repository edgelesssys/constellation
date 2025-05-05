terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.27.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.3.0"
    }
  }
}

# Configure Azure resource management provider
provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
  subscription_id = var.subscription_id
  # This enables all resource providers.
  # In the future, we might want to use `resource_providers_to_register` to registers just the ones we need.
  resource_provider_registrations = "all"
}

# Configure Azure active directory provider
provider "azuread" {
  tenant_id = data.azurerm_subscription.current.tenant_id
}

# Access current subscription (available via Azure CLI)
data "azurerm_subscription" "current" {}

# Access current AzureAD configuration
data "azuread_client_config" "current" {}

# Create base resource group
resource "azurerm_resource_group" "base_resource_group" {
  name     = var.resource_group_name
  location = var.location
}

# Create identity resource group
resource "azurerm_resource_group" "identity_resource_group" {
  name     = "${var.resource_group_name}-identity"
  location = var.location
}

# Create managed identity
resource "azurerm_user_assigned_identity" "identity_uami" {
  location            = var.location
  name                = var.service_principal_name
  resource_group_name = azurerm_resource_group.identity_resource_group.name
}

# Assign roles to managed identity
resource "azurerm_role_assignment" "virtual_machine_contributor_role" {
  scope                = azurerm_resource_group.base_resource_group.id
  role_definition_name = "Virtual Machine Contributor"
  principal_id         = azurerm_user_assigned_identity.identity_uami.principal_id
}

resource "azurerm_role_assignment" "application_insights_component_contributor_role" {
  scope                = azurerm_resource_group.base_resource_group.id
  role_definition_name = "Application Insights Component Contributor"
  principal_id         = azurerm_user_assigned_identity.identity_uami.principal_id
}

resource "azurerm_role_assignment" "uami_owner_role" {
  scope                = azurerm_resource_group.base_resource_group.id
  role_definition_name = "Owner"
  principal_id         = azurerm_user_assigned_identity.identity_uami.principal_id
}
