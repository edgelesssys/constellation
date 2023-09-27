variable "name" {
  type        = string
  description = "Base name of the cluster."
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

variable "location" {
  type        = string
  description = "The Azure location to deploy the cluster in."
}

variable "image_id" {
  type        = string
  description = "The image to use for the cluster nodes."
}

variable "create_maa" {
  type        = bool
  default     = false
  description = "Whether to create a Microsoft Azure attestation provider."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
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
  description = "The name of the Azure resource group to create the Constellation cluster in."
}
variable "user_assigned_identity" {
  type        = string
  description = "The name of the user assigned identity to attach to the nodes of the cluster. Should be of format: /subscriptions/$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME"
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes apiserver. If not set, the default endpoint will be used."
}

variable "internal_load_balancer" {
  type        = bool
  default     = false
  description = "Whether to use an internal load balancer for the Constellation."
}
