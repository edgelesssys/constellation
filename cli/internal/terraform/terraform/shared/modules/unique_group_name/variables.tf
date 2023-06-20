variable "node_group_name" {
  type        = string
  description = "Name of the instance group."
}

variable "base_name" {
  type        = string
  description = "Base name of cluster."
}


variable "role" {
  type        = string
  description = "The role of the group (ControlPlane|Worker)."
}
