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
  type        = string
  description = "role of the node in the constellation. either 'control-plane' or 'worker'"
}

variable "machine" {
  type        = string
  description = "machine type. use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'"
}

variable "name" {
  type        = string
  description = "name prefix of the cluster VMs"
}
