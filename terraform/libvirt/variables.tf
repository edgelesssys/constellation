variable "constellation_coreos_image_qcow2" {
  type        = string
  description = "constellation OS qcow file path"
}

variable "control_plane_count" {
  type        = number
  default     = 3
  description = "amount of control plane nodes"
}

variable "worker_count" {
  type        = number
  default     = 2
  description = "amount of worker nodes"
}

variable "vcpus" {
  type        = number
  default     = 2
  description = "amount of vcpus per instance"
}

variable "memory" {
  type        = number
  default     = 2048
  description = "amount of memory per instance (MiB)"
}

variable "state_disk_size" {
  type        = number
  default     = 10
  description = "size of state disk (GiB)"
}

variable "ip_range_start" {
  type        = number
  default     = 100
  description = "first ip address to use within subnet"
}
