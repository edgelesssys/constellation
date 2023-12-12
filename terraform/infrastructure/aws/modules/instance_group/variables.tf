variable "base_name" {
  type        = string
  description = "Base name of the instance group."
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

variable "uid" {
  type        = string
  description = "Unique Identifier (UID) of the cluster."
}

variable "instance_type" {
  type        = string
  description = "Instance type for the nodes."
}

variable "initial_count" {
  type        = number
  description = "Number of instances in the instance group."
}

variable "image_id" {
  type        = string
  description = "Amazon Machine Image (AMI) ID for the cluster's nodes."
}

variable "state_disk_type" {
  type        = string
  description = "EBS disk type for the state disk of the nodes."
}

variable "state_disk_size" {
  type        = number
  description = "Disk size for the state disk of the nodes [GB]."
}

variable "target_group_arns" {
  type        = list(string)
  description = "ARN of the target group."
}

variable "subnetwork" {
  type        = string
  description = "Name of the subnetwork to use."
}

variable "iam_instance_profile" {
  type        = string
  description = "IAM instance profile for the nodes."
}

variable "security_groups" {
  type        = list(string)
  description = "List of security group IDs for an instance."
}

variable "tags" {
  type        = map(string)
  description = "Tags to add to the instance group."
}

variable "enable_snp" {
  type        = bool
  default     = true
  description = "Enable AMD SEV-SNP for the instances."
}

variable "zone" {
  type        = string
  description = "Zone to deploy the instance group in."
}
