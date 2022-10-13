variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "instance_count" {
  type        = number
  description = "The number of instances in this scale set."
}

variable "instance_type" {
  type        = string
  description = "The Azure instance type to deploy."
}

variable "state_disk_size" {
  type        = number
  default     = 30
  description = "The size of the state disk in GB."
}

variable "resource_group" {
  type        = string
  description = "The name of the Azure resource group to create the Constellation cluster in."
}

variable "location" {
  type        = string
  description = "The Azure location to deploy the cluster in."
}

variable "image_id" {
  type        = string
  description = "The image to use for the cluster nodes."
}

variable "user_assigned_identity" {
  type        = string
  description = "The name of the user assigned identity to attache to the nodes of the cluster."
}

variable "state_disk_type" {
  type        = string
  default     = "Premium_LRS"
  description = "The type of the state disk."
}

variable "network_security_group_id" {
  type        = string
  description = "The ID of the network security group to use for the scale set."
}

variable "backend_address_pool_ids" {
  type        = list(string)
  description = "The IDs of the backend address pools to use for the scale set."
}

variable "subnet_id" {
  type        = string
  description = "The ID of the subnet to use for the scale set."
}

variable "tags" {
  type        = map(string)
  description = "The tags to add to the scale set."
}

variable "confidential_vm" {
  type        = bool
  default     = true
  description = "Whether to deploy the cluster nodes as confidential VMs."
}
