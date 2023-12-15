variable "name" {
  type        = string
  description = "Base name of the load balancer."
}

variable "region" {
  type        = string
  description = "Region to create the load balancer in."
}

variable "network" {
  type        = string
  description = "Network to which network resources will be attached."
}

variable "backend_subnet" {
  type        = string
  description = "Subnet to which backend network resources will be attached."
}

variable "health_check" {
  type        = string
  description = "Type of the health check. Can either be 'HTTPS' or 'TCP'."
  validation {
    condition     = contains(["HTTPS", "TCP"], var.health_check)
    error_message = "Health check must be either 'HTTPS' or 'TCP'."
  }
}

variable "port" {
  type        = string
  description = "Port to listen on for incoming traffic."
}

variable "backend_port_name" {
  type        = string
  description = "Name of the load balancer's backend port. The same name should appear in the instance groups referenced by this service."
}

variable "backend_instance_group" {
  type        = string
  description = "Full URL of the instance group resource from which the load balancer will direct traffic."
}

variable "ip_address" {
  type        = string
  description = "IP address that this forwarding rule serves."
}

variable "frontend_labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the forwarding rule."
}
