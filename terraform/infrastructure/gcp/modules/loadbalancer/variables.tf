variable "name" {
  type        = string
  description = "Base name of the load balancer."
}

variable "health_check" {
  type        = string
  description = "The type of the health check. 'HTTPS' or 'TCP'."
}

variable "backend_port_name" {
  type        = string
  description = "Name of backend port. The same name should appear in the instance groups referenced by this service."
}

variable "backend_instance_groups" {
  type        = list(string)
  description = "The URLs of the instance group resources from which the load balancer will direct traffic."
}

variable "ip_address" {
  type        = string
  description = "The IP address that this forwarding rule serves. An address can be specified either by a literal IP address or a reference to an existing Address resource."
}

variable "port" {
  type        = number
  description = "The port on which to listen for incoming traffic."
}

variable "frontend_labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the forwarding rule."
}
