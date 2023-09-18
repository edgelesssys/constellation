variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = number // number of instances in the node group
    disk_size     = number // size of state disk (GiB)
    vcpus         = number
    memory        = number // amount of memory per instance (MiB)
  }))
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }

  description = "A map of node group names to node group configurations."
}

variable "machine" {
  type        = string
  default     = "q35"
  description = "machine type. use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'"
}

variable "libvirt_uri" {
  type        = string
  description = "libvirt socket uri"
}

variable "constellation_boot_mode" {
  type        = string
  description = "constellation boot mode. Can be 'uefi' or 'direct-linux-boot'"
  validation {
    condition = anytrue([
      var.constellation_boot_mode == "uefi",
      var.constellation_boot_mode == "direct-linux-boot",
    ])
    error_message = "constellation_boot_mode must be 'uefi' or 'direct-linux-boot'"
  }
}

variable "constellation_os_image" {
  type        = string
  description = "constellation OS file path"
}

variable "constellation_kernel" {
  type        = string
  description = "constellation Kernel file path"
  default     = ""
}

variable "constellation_initrd" {
  type        = string
  description = "constellation initrd file path"
  default     = ""
}

variable "constellation_cmdline" {
  type        = string
  description = "constellation kernel cmdline"
  default     = ""
}

variable "image_format" {
  type        = string
  default     = "qcow2"
  description = "image format"
}
variable "firmware" {
  type        = string
  default     = "/usr/share/OVMF/OVMF_CODE.secboot.fd"
  description = "path to UEFI firmware file. Use \"OVMF_CODE_4M.ms.fd\" on Ubuntu and \"OVMF_CODE.fd\" or \"OVMF_CODE.secboot.fd\" on Fedora."
}

variable "nvram" {
  type        = string
  description = "path to UEFI NVRAM template file. Used for secure boot."
}

variable "metadata_api_image" {
  type        = string
  description = "container image of the QEMU metadata api server"
}

variable "metadata_libvirt_uri" {
  type        = string
  description = "libvirt uri for the metadata api server"
}

variable "libvirt_socket_path" {
  type        = string
  description = "path to libvirt socket in case of unix socket"
}

variable "name" {
  type        = string
  default     = "constellation"
  description = "name prefix of the cluster VMs"
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes apiserver. If not set, the default endpoint will be used."
}
