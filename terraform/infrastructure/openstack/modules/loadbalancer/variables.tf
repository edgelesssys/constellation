variable "name" {
  type        = string
  description = "Base name of the load balancer."
}

variable "member_ips" {
  type        = list(string)
  description = "IP addresses of the members of the load balancer pool."
  default     = []
}

variable "loadbalancer_id" {
  type        = string
  description = "ID of the load balancer."
}

variable "subnet_id" {
  type        = string
  description = "ID of the members subnet."
}

variable "port" {
  type        = number
  description = "Port to listen on incoming traffic."
}
