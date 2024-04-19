variable "base_name" {
  description = "Base name of the jump host."
  type        = string
}

variable "ports" {
  description = "Ports to forward to the load balancer."
  type        = list(number)
}

variable "subnet_id" {
  description = "Subnet ID to deploy the jump host into."
  type        = string
}

variable "lb_internal_ip" {
  description = "Internal IP of the load balancer."
  type        = string
}

variable "resource_group" {
  description = "Resource group name to deploy the jump host into."
  type        = string
}

variable "location" {
  description = "Location to deploy the jump host into."
  type        = string
}

variable "tags" {
  description = "Tags of the jump host."
  type        = map(any)
}
