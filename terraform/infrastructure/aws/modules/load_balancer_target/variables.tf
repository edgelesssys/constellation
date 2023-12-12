variable "base_name" {
  type        = string
  description = "Base name of the load balancer target."
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

variable "healthcheck_protocol" {
  type        = string
  default     = "TCP"
  description = "Type of the load balancer target."
}

variable "healthcheck_path" {
  type        = string
  default     = ""
  description = "Path for health check."
}

variable "tags" {
  type        = map(string)
  description = "Tags to add to the loadbalancer."
}
