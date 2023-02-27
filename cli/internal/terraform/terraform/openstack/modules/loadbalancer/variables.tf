variable "name" {
  type        = string
  description = "Base name of the load balancer rule."
}

variable "member_ips" {
  type        = list(string)
  description = "The IP addresses of the members of the load balancer pool."
  default     = []
}

variable "loadbalancer_id" {
  type        = string
  description = "The ID of the load balancer."
}

variable "subnet_id" {
  type        = string
  description = "The ID of the members subnet."
}

variable "port" {
  type        = number
  description = "The port on which to listen for incoming traffic."
}
