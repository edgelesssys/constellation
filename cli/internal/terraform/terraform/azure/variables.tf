variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "control_plane_count" {
  type        = number
  description = "The number of control plane nodes to deploy."
}

variable "worker_count" {
  type        = number
  description = "The number of worker nodes to deploy."
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

variable "user_assigned_identity" {
  type        = string
  description = "The name of the user assigned identity to attache to the nodes of the cluster."
}

variable "instance_type" {
  type        = string
  description = "The Azure instance type to deploy."
}

variable "state_disk_type" {
  type        = string
  default     = "Premium_LRS"
  description = "The type of the state disk."
}

variable "image_id" {
  type        = string
  description = "The image to use for the cluster nodes."
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
