variable "base_name" {
  type        = string
  description = "Base name of the scale set."
}

variable "node_group_name" {
  type        = string
  description = "Constellation name for the node group (used for configuration and CSP-independent naming)."
}

variable "role" {
  type        = string
  description = "Role of the instance group."
  validation {
    condition     = contains(["control-plane", "worker"], var.role)
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "tags" {
  type        = map(string)
  description = "Tags to include in the scale set."
}

variable "zones" {
  type        = list(string)
  description = "List of availability zones."
  default     = null
}

variable "initial_count" {
  type        = number
  description = "Number of instances in this scale set."
}

variable "instance_type" {
  type        = string
  description = "Azure instance type to deploy."
}

variable "state_disk_size" {
  type        = number
  default     = 30
  description = "Disk size for the state disk of the nodes [GB]."
}

variable "resource_group" {
  type        = string
  description = "Name of the Azure resource group to create the Constellation cluster in."
}

variable "location" {
  type        = string
  description = "Azure location to deploy the cluster in."
}

variable "image_id" {
  type        = string
  description = "OS Image reference for the cluster's nodes."
}

variable "user_assigned_identity" {
  type        = string
  description = "Name of the user assigned identity to attache to the nodes of the cluster."
}

variable "state_disk_type" {
  type        = string
  default     = "Premium_LRS"
  description = "Type of the state disk."
}

variable "network_security_group_id" {
  type        = string
  description = "ID of the network security group to use for the scale set."
}

variable "backend_address_pool_ids" {
  type        = list(string)
  description = "IDs of the backend address pools to use for the scale set."
}

variable "subnet_id" {
  type        = string
  description = "ID of the subnet to use for the scale set."
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

variable "marketplace_image" {
  type = object({
    name      = string
    publisher = string
    product   = string
    version   = string
  })
  default     = null
  description = "Marketplace image to use for the cluster nodes."
}
