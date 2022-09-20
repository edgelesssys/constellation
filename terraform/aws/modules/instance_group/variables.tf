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

variable "disk_size" {
  type        = number
  description = "Disk size for the nodes, in GB."
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
