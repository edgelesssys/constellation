variable "base_name" {
  type        = string
  description = "Base name of the jump host."
}

variable "labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the jump host."
}

variable "subnetwork" {
  type        = string
  description = "Subnetwork to deplyo the jump host into."
}

variable "zone" {
  type        = string
  description = "Zone to deploy the jump host into."
}

variable "lb_internal_ip" {
  type        = string
  description = "Internal IP of the load balancer."
}

variable "ports" {
  type        = list(number)
  description = "Ports to forward to the load balancer."
}
