variable "name" {
  type        = string
  description = "Base name of the instance group."
}

variable "role" {
  type        = string
  description = "The role of the instance group. Has to be 'ControlPlane' or 'Worker'."
}

variable "uid" {
  type        = string
  description = "UID of the cluster. This is used for tags."
}

variable "instance_type" {
  type        = string
  description = "Instance type for the nodes."
}

variable "instance_count" {
  type        = number
  description = "Number of instances in the instance group."
}

variable "image_id" {
  type        = string
  description = "Image ID for the nodes."
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
  description = "List of IDs of the security groups for an instance."
}

variable "tags" {
  type        = map(string)
  description = "The tags to add to the instance group."
}
