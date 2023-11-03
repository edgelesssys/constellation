variable "base_name" {
  description = "Base name of the jump host"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID to deploy the jump host into"
  type        = string
}

variable "lb_internal_ip" {
  description = "Internal IP of the load balancer"
  type        = string
}

variable "iam_instance_profile" {
  description = "IAM instance profile to attach to the jump host"
  type        = string
}

variable "ports" {
  description = "Ports to forward to the load balancer"
  type        = list(number)
}

variable "security_group_id" {
  description = "Security group to attach to the jump host"
  type        = string
}
