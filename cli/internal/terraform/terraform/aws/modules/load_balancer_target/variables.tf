variable "name" {
  type        = string
  description = "Name of the load balancer target."
}

variable "port" {
  type        = string
  description = "Port of the load balancer target."
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC."
}

variable "lb_arn" {
  type        = string
  description = "ARN of the load balancer."
}
