variable "name" {
  type        = string
  description = "Base name of the cluster."
}

variable "node_groups" {
  type = map(object({
    role         =  string
    instance_count = number
    instance_type = string
    confidential_vm = bool
    secure_boot = bool
    user_assigned_identity = string
    disk_size     = number
    disk_type     = string
    resource_group = string
    location     = string
  }))
  description = "A map of node group names to node group configurations."
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
