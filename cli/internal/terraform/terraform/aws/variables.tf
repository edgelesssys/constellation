variable "name" {
  type        = string
  description = "Name of your Constellation"
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
  description = "EBS disk type for the state disk of the nodes"
}

variable "state_disk_size" {
  type        = number
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
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}
