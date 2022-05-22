variable "amount" {
  type        = number
  description = "amount of nodes"
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

variable "cidr" {
  type        = string
  description = "subnet to use for dhcp"
}

variable "network_id" {
  type        = string
  description = "id of the network to use"
}

variable "pool" {
  type        = string
  description = "name of the storage pool to use"
}

variable "boot_volume_id" {
  type        = string
  description = "id of the constellation boot disk"
}

variable "role" {
  type = string
}
