variable "name" {
  type        = string
  description = "Base name of the load balancer."
}

variable "member_ips" {
  type        = list(string)
  description = "IP addresses of the members of the load balancer pool."
  default     = []
}

variable "network_id" {
  type        = string
  description = "ID of the network."
}

variable "external_address" {
  type        = string
  description = "External address of the load balancer."
}

variable "ports" {
  type        = map(number)
  description = "Ports to listen on incoming traffic."
}

variable "stackit_project_id" {
  type        = string
  description = "STACKIT project ID."
}
