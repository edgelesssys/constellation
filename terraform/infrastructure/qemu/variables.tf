# Variables common to all CSPs

variable "name" {
  type        = string
  default     = "constellation"
  description = "Name of the Constellation cluster."
}

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

variable "image_id" {
  type        = string
  description = "Path to the OS image for the cluster's nodes."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes API server. If not set, the default endpoint will be used."
}

# QEMU-specific variables

variable "machine" {
  type        = string
  default     = "q35"
  description = "Machine type. Use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'."
}

variable "libvirt_uri" {
  type        = string
  description = "URI of the Libvirt socket."
}

variable "constellation_boot_mode" {
  type        = string
  description = "Constellation boot mode. Can be 'uefi' or 'direct-linux-boot'."
  validation {
    condition = anytrue([
      var.constellation_boot_mode == "uefi",
      var.constellation_boot_mode == "direct-linux-boot",
    ])
    error_message = "constellation_boot_mode must be 'uefi' or 'direct-linux-boot'."
  }
}

variable "constellation_kernel" {
  type        = string
  description = "Constellation Kernel file path."
  default     = ""
}

variable "constellation_initrd" {
  type        = string
  description = "Constellation initrd file path."
  default     = ""
}

variable "constellation_cmdline" {
  type        = string
  description = "Constellation kernel cmdline."
  default     = ""
}

variable "image_format" {
  type        = string
  default     = "qcow2"
  description = "Image format."
}

variable "firmware" {
  type        = string
  default     = "/usr/share/OVMF/OVMF_CODE.fd"
  description = "Path to UEFI firmware file. Use \"OVMF_CODE_4M.ms.fd\" on Ubuntu and \"OVMF_CODE.fd\" or \"OVMF_CODE.secboot.fd\" on Fedora."
}

variable "nvram" {
  type        = string
  description = "Path to UEFI NVRAM template file. Used for secure boot."
}

variable "metadata_api_image" {
  type        = string
  description = "Container image of the QEMU metadata API server."
}

variable "metadata_libvirt_uri" {
  type        = string
  description = "Libvirt URI for the metadata API server."
}

variable "libvirt_socket_path" {
  type        = string
  description = "Path to Libvirt socket in case of unix socket."
}
