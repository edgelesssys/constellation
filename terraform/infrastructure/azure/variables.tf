# Variables common to all CSPs

variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
}

variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = optional(number)
    instance_type = string
    disk_size     = number
    disk_type     = string
    zones         = optional(list(string))
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "image_id" {
  type        = string
  description = "OS image reference for the cluster's nodes."
}

variable "debug" {
  type        = bool
  default     = false
  description = "DO NOT USE IN PRODUCTION. Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes API server. If not set, the default endpoint will be used."
}

variable "internal_load_balancer" {
  type        = bool
  default     = false
  description = "Whether to use an internal load balancer for the cluster."
}

# Azure-specific variables

variable "subscription_id" {
  type        = string
  description = "Azure subscription ID. This can also be sourced from the ARM_SUBSCRIPTION_ID environment variable: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#subscription_id"
  default     = ""
}

variable "location" {
  type        = string
  description = "Azure location to deploy the cluster in."
}

variable "create_maa" {
  type        = bool
  default     = false
  description = "Whether to create a Microsoft Azure Attestation (MAA) provider."
}

variable "confidential_vm" {
  type        = bool
  default     = true
  description = "Whether to deploy the cluster nodes as confidential VMs."
}

variable "secure_boot" {
  type        = bool
  default     = false
  description = "Whether to deploy the cluster nodes with secure boot."
}

variable "resource_group" {
  type        = string
  description = "Name of the Azure resource group to create the cluster in."
}

variable "user_assigned_identity" {
  type        = string
  description = "Name of the user assigned identity to attach to the nodes of the cluster. Should be of format: /subscriptions/$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME"
}

variable "marketplace_image" {
  type = object({
    name      = string
    publisher = string
    product   = string
    version   = string
  })
  default     = null
  description = "Marketplace image for the cluster's nodes."
}

variable "additional_tags" {
  type        = map(any)
  default     = {}
  description = "Additional tags that should be applied to created resources."
}
