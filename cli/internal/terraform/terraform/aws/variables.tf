variable "name" {
  type        = string
  default     = "constell"
  description = "Name of your Constellation"
  validation {
    condition     = length(var.name) < 10
    error_message = "The name of the Constellation must be shorter than 10 characters"
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

variable "instance_type" {
  type        = string
  description = "Instance type for worker nodes"
}

variable "state_disk_type" {
  type        = string
  default     = "gp2"
  description = "EBS disk type for the state disk of the nodes"
}

variable "state_disk_size" {
  type        = number
  default     = 30
  description = "Disk size for the state disk of the nodes [GB]"
}

variable "control_plane_count" {
  type        = number
  description = "Number of control plane nodes"
}

variable "worker_count" {
  type        = number
  description = "Number of worker nodes"
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
