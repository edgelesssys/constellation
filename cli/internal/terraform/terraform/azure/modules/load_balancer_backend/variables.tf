variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "frontend_ip_configuration_name" {
  type        = string
  description = "The name of the frontend IP configuration to use for the load balancer."
}

variable "loadbalancer_id" {
  type        = string
  description = "The ID of the load balancer to add the backend to."
}

variable "ports" {
  type = list(object({
    name                  = string
    port                  = number
    health_check_protocol = string
    path                  = string
  }))
  description = "The ports to add to the backend. Protocol can be either 'Tcp' or 'Https'. Path is only used for 'Https' protocol and can otherwise be null."
}
