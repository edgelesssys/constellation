variable "constellation_coreos_image" {
  type        = string
  description = "constellation OS file path"
}

variable "image_format" {
  type        = string
  default     = "qcow2"
  description = "image format"
}

variable "control_plane_count" {
  type        = number
  description = "amount of control plane nodes"
}

variable "worker_count" {
  type        = number
  description = "amount of worker nodes"
}

variable "vcpus" {
  type        = number
  description = "amount of vcpus per instance"
}

variable "memory" {
  type        = number
  description = "amount of memory per instance (MiB)"
}

variable "state_disk_size" {
  type        = number
  description = "size of state disk (GiB)"
}

variable "ip_range_start" {
  type        = number
  description = "first ip address to use within subnet"
}

variable "machine" {
  type        = string
  default     = "q35"
  description = "machine type. use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'"
}

variable "metadata_api_image" {
  type        = string
  description = "container image of the QEMU metadata api server"
}

variable "name" {
  type        = string
  default     = "constellation"
  description = "name prefix of the cluster VMs"
}
