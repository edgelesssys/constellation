variable "name" {
  type        = string
  description = "Name of your Constellation"
  validation {
    condition     = length(var.name) < 10
    error_message = "The name of the Constellation must be shorter than 10 characters"
  }
}

variable "node_groups" {
  type = map(object({
    role           = string
    instance_count = optional(number)
    instance_type  = string
    disk_size      = number
    disk_type      = string
    zone           = string
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "iam_instance_profile_worker_nodes" {
  type        = string
  description = "Name of the IAM instance profile for worker nodes"
}

variable "iam_instance_profile_control_plane" {
  type        = string
  description = "Name of the IAM instance profile for control plane nodes"
}

variable "ami" {
  type        = string
  description = "AMI ID"
  validation {
    condition     = length(var.ami) > 4 && substr(var.ami, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
}

variable "region" {
  type        = string
  description = "The AWS region to create the cluster in"
}

variable "zone" {
  type        = string
  description = "The AWS availability zone name to create the cluster in"
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "enable_snp" {
  type        = bool
  default     = true
  description = "Enable AMD SEV SNP. Setting this to true sets the cpu-option AmdSevSnp to enable."
}
