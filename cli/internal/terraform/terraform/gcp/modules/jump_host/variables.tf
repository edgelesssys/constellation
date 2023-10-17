variable "base_name" {
  type        = string
  description = "Base name of the instance group."
}

variable "labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the instance group."
}

variable "subnetwork" {
  type        = string
  description = "Name of the subnetwork to use."
}

variable "zone" {
  type        = string
  description = "Zone to deploy the instance group in."
}

variable "lb_internal_ip" {
  type        = string
  description = "Internal IP of the load balancer."
}

variable "ports" {
  type        = list(number)
  description = "Ports to forward to the load balancer."
}
