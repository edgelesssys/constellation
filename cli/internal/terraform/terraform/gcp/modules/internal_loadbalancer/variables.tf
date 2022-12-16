variable "name" {
  type        = string
  description = "Base name of the load balancer."
}

variable "region" {
  type        = string
  description = "The region where the load balancer will be created."
}

variable "network" {
  type        = string
  description = "The network to which all network resources will be attached."
}

variable "backend_subnet" {
  type        = string
  description = "The subnet to which all backend network resources will be attached."
}

variable "health_check" {
  type        = string
  description = "The type of the health check. 'HTTPS' or 'TCP'."
}

variable "port" {
  type        = string
  description = "The port on which to listen for incoming traffic."
}

variable "port_name" {
  type        = string
  description = "Name of backend port. The same name should appear in the instance groups referenced by this service."
}

variable "backend_instance_group" {
  type        = string
  description = "The URL of the instance group resource from which the load balancer will direct traffic."
}

variable "ip_address" {
  type        = string
  description = "The IP address that this forwarding rule serves."
}

variable "frontend_labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the forwarding rule."
}
