variable "name" {
  type = string
}

variable "worker_nodes_iam_instance_profile" {
  type        = string
  description = "Name of the IAM instance profile for worker nodes"
}

variable "control_plane_iam_instance_profile" {
  type        = string
  description = "Name of the IAM instance profile for control plane nodes"
}
