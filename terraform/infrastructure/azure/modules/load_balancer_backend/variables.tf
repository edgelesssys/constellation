variable "name" {
  type        = string
  default     = "constell"
  description = "Name of the Constellation cluster."
}

variable "frontend_ip_configuration_name" {
  type        = string
  description = "Name of the frontend IP configuration to use for the load balancer."
}

variable "loadbalancer_id" {
  type        = string
  description = "ID of the load balancer to add the backend to."
}

variable "ports" {
  type = list(object({
    name                  = string
    port                  = number
    health_check_protocol = string
    path                  = string
  }))
  description = "Ports to add to the backend. Healtch check protocol can be either 'Tcp' or 'Https'. Path is only used for the 'Https' protocol and can otherwise be null."
}
