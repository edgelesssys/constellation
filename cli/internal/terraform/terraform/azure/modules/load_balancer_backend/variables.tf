variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "loadbalancer_id" {
  type        = string
  description = "The ID of the load balancer to add the backend to."
}

variable "ports" {
  type = list(object({
    name = string
    port = number
  }))
  description = "The ports to add to the backend."
}
