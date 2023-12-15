variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
}

variable "health_check" {
  type        = string
  description = "Type of the health check. Can either be 'HTTPS' or 'TCP'."
}

variable "backend_port_name" {
  type        = string
  description = "Name of the load balancer's backend port. The same name should appear in the instance groups referenced by this service."
}

variable "backend_instance_groups" {
  type        = list(string)
  description = "URLs of the instance group resources from which the load balancer will direct traffic."
}

variable "ip_address" {
  type        = string
  description = "IP address that this forwarding rule serves. An address can be specified either by a literal IP address or a reference to an existing Address resource."
}

variable "port" {
  type        = number
  description = "Port to listen on for incoming traffic."
}

variable "frontend_labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the load balancer's forwarding rule."
}
