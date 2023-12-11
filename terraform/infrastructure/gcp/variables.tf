variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "node_groups" {
  type = map(object({
    role          = string
    zone          = string
    instance_type = string
    disk_size     = number
    disk_type     = string
    initial_count = number
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "project" {
  type        = string
  description = "The GCP project to deploy the cluster in."
}

variable "region" {
  type        = string
  description = "The GCP region to deploy the cluster in."
}

variable "zone" {
  type        = string
  description = "The GCP zone to deploy the cluster in."
}

variable "image_id" {
  type        = string
  description = "The GCP image reference to use for the cluster nodes."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes apiserver. If not set, the default endpoint will be used."
}

variable "internal_load_balancer" {
  type        = bool
  default     = false
  description = "Enable internal load balancer. This can only be enabled if the control-plane is deployed in one zone."
}
