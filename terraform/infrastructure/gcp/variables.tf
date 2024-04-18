# Variables common to all CSPs

variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
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

# GCP-specific variables

variable "project" {
  type        = string
  description = "GCP project to deploy the cluster in."
}

variable "region" {
  type        = string
  description = "GCP region to deploy the cluster in."
}

variable "zone" {
  type        = string
  description = "GCP zone to deploy the cluster in."
}

variable "cc_technology" {
  type        = string
  description = "The confidential computing technology to use for the nodes. One of `SEV`, `SEV_SNP`."
  validation {
    condition     = contains(["SEV", "SEV_SNP"], var.cc_technology)
    error_message = "The confidential computing technology has to be 'SEV' or 'SEV_SNP'."
  }
}

variable "additional_labels" {
  type = map
  description = "Additional labels that should be given to created recources."
}
