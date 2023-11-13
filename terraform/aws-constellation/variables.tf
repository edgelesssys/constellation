variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
}

variable "image" {
  type        = string
  description = "Node image reference or semantic release version. When not set, the latest default version will be used."
  default     = "@@CONSTELLATION_VERSION@@"
}

variable "microservice_version" {
  type        = string
  description = "Microservice version. When not set, the latest default version will be used."
  default     = ""
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version. When not set, the latest default version will be used."
  default     = ""
}

variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = optional(number)
    instance_type = string
    disk_size     = number
    disk_type     = string
    zone          = string
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "zone" {
  type        = string
  description = "The AWS availability zone name to create the cluster in."
}

variable "debug" {
  type        = bool
  default     = false
  description = "DON'T USE IN PRODUCTION: Enable debug mode and allow the use of debug images."
}

variable "enable_snp" {
  type        = bool
  default     = true
  description = "Enable AMD SEV-SNP."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint (DNS Name) to use for the Constellation API server. If not set, the default endpoint will be used."
}

variable "internal_load_balancer" {
  type        = bool
  default     = false
  description = "Use an internal load balancer."
}

variable "name_prefix" {
  type        = string
  description = "Prefix for all resources."
}
