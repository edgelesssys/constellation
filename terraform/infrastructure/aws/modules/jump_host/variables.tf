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

variable "iam_instance_profile" {
  description = "IAM instance profile to attach to the jump host."
  type        = string
}

variable "security_groups" {
  type        = list(string)
  description = "List of IDs of the security groups for an instance."
}

variable "additional_tags" {
  type        = map(any)
  description = "Additional tags for the jump host."
}
