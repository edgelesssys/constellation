# Variables common to all CSPs

variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
  validation {
    condition     = length(var.name) <= 10
    error_message = "The length of the name of the Constellation must be <= 10 characters."
  }
  validation {
    condition     = var.name == lower(var.name)
    error_message = "The name of the Constellation must be in lowercase."
  }
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

variable "image_id" {
  type        = string
  description = "Amazon Machine Image (AMI) ID for the cluster's nodes."
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

# AWS-specific variables

variable "iam_instance_profile_name_worker_nodes" {
  type        = string
  description = "Name of the IAM instance profile for worker nodes."
}

variable "iam_instance_profile_name_control_plane" {
  type        = string
  description = "Name of the IAM instance profile for control plane nodes."
}

variable "region" {
  type        = string
  description = "AWS region to create the cluster in."
}

variable "zone" {
  type        = string
  description = "AWS availability zone name to create the cluster in."
}

variable "enable_snp" {
  type        = bool
  default     = true
  description = "Enable AMD SEV SNP. Setting this to true sets the cpu-option AmdSevSnp to enable."
}

variable "additional_tags" {
  type        = map(any)
  default     = {}
  description = "Additional tags that should be applied to created resources."
}
