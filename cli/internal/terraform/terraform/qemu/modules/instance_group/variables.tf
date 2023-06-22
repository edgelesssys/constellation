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

variable "boot_mode" {
  type        = string
  description = "boot mode. Can be 'uefi' or 'direct-linux-boot'"
  validation {
    condition     = can(regex("^(uefi|direct-linux-boot)$", var.boot_mode))
    error_message = "boot_mode must be 'uefi' or 'direct-linux-boot'"
  }
}

variable "boot_volume_id" {
  type        = string
  description = "id of the constellation boot disk"
}

variable "kernel_volume_id" {
  type        = string
  description = "id of the constellation kernel volume"
  default     = ""
}

variable "initrd_volume_id" {
  type        = string
  description = "id of the constellation initrd volume"
  default     = ""
}

variable "kernel_cmdline" {
  type        = string
  description = "kernel cmdline"
  default     = ""
}

variable "role" {
  type        = string
  description = "role of the node in the constellation. either 'control-plane' or 'worker'"
}

variable "machine" {
  type        = string
  description = "machine type. use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'"
}

variable "firmware" {
  type        = string
  description = "path to UEFI firmware file. Ignored for direct-linux-boot."
}

variable "nvram" {
  type        = string
  description = "path to UEFI NVRAM template file. Used for secure boot."
}
variable "base_name" {
  type        = string
  description = "name prefix of the cluster VMs"
}

variable "node_group_name" {
  type        = string
  description = "name of the node group"
}
