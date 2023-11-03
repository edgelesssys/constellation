terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.74.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "2.43.0"
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
}

# Configure Azure active directory provider
provider "azuread" {
  tenant_id = data.azurerm_subscription.current.tenant_id
}

# Access current subscription (available via Azure CLI)
data "azurerm_subscription" "current" {}

# # Access current AzureAD configuration
data "azuread_client_config" "current" {}

# Create base resource group
resource "azurerm_resource_group" "base_resource_group" {
  name     = var.resource_group_name
  location = var.region
}

# Create identity resource group
resource "azurerm_resource_group" "identity_resource_group" {
  name     = "${var.resource_group_name}-identity"
  location = var.region
}

# Create managed identity
resource "azurerm_user_assigned_identity" "identity_uami" {
  location            = var.region
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
