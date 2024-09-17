variable "subscription_id" {
  type        = string
  description = "Azure subscription ID. This can also be sourced from the ARM_SUBSCRIPTION_ID environment variable: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#subscription_id"
  default     = ""
}

variable "resource_group_name" {
  type        = string
  description = "Name for the resource group the cluster should reside in."
}

variable "service_principal_name" {
  type        = string
  description = "Name for the service principal used within the cluster."
}

variable "location" {
  type        = string
  description = "Azure location the cluster should reside in."
}
