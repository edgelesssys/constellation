variable "name" {
  type        = string
  description = "Name of the load balancer."
}

variable "port" {
  type        = string
  description = "Port of the load balancer."
}

variable "vpc" {
  type        = string
  description = "ID of the VPC."
}

variable "subnet" {
  type        = string
  description = "ID of the subnets."
}
