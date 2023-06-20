variable "base_name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "node_group_name" {
  type        = string
  description = "Name of the node group."
}

variable "role" {
  type        = string
  description = "The role of the group (ControlPlane|Worker)."
}

variable "loadbalancer_id" {
  type        = string
  description = "The ID of the load balancer to add the backend to."
}

variable "ports" {
  type = list(object({
    name     = string
    port     = number
    protocol = string
    path     = string
  }))
  description = "The ports to add to the backend. Protocol can be either 'Tcp' or 'Https'. Path is only used for 'Https' protocol and can otherwise be null."
}
