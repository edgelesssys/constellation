variable "base_name" {
  type        = string
  description = "Name prefix fort the cluster's VMs."
}

variable "node_group_name" {
  type        = string
  description = "Constellation name for the node group (used for configuration and CSP-independent naming)."
}

variable "amount" {
  type        = number
  description = "Amount of nodes."
}

variable "vcpus" {
  type        = number
  description = "Amount of vCPUs per instance."
}

variable "memory" {
  type        = number
  description = "Amount of memory per instance (MiB)."
}

variable "state_disk_size" {
  type        = number
  description = "Disk size for the state disk of the nodes [GB]."
}

variable "cidr" {
  type        = string
  description = "Subnet to use for DHCP."
}

variable "network_id" {
  type        = string
  description = "ID of the Libvirt network to use."
}

variable "pool" {
  type        = string
  description = "Name of the Libvirt storage pool to use."
}

variable "boot_mode" {
  type        = string
  description = "Boot mode. Can be 'uefi' or 'direct-linux-boot'"
  validation {
    condition     = can(regex("^(uefi|direct-linux-boot)$", var.boot_mode))
    error_message = "boot_mode must be 'uefi' or 'direct-linux-boot'"
  }
}

variable "boot_volume_id" {
  type        = string
  description = "ID of the Constellation boot disk."
}

variable "kernel_volume_id" {
  type        = string
  description = "ID of the Constellation kernel volume."
  default     = ""
}

variable "initrd_volume_id" {
  type        = string
  description = "ID of the constellation initrd volume."
  default     = ""
}

variable "kernel_cmdline" {
  type        = string
  description = "Kernel cmdline."
  default     = ""
}

variable "role" {
  type        = string
  description = "Role of the node in the Constellation cluster. Can either  be'control-plane' or 'worker'."
}

variable "machine" {
  type        = string
  description = "Machine type. Use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'."
}

variable "firmware" {
  type        = string
  description = "Path to UEFI firmware file. Ignored for direct-linux-boot."
}

variable "nvram" {
  type        = string
  description = "Path to UEFI NVRAM template file. Used for secure boot."
}
