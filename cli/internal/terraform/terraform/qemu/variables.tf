variable "libvirt_uri" {
  type        = string
  description = "libvirt socket uri"
}

variable "constellation_os_image" {
  type        = string
  description = "constellation OS file path"
}

variable "constellation_kernel" {
  type        = string
  description = "constellation Kernel file path"
}

variable "constellation_initrd" {
  type        = string
  description = "constellation initrd file path"
}

variable "constellation_cmdline" {
  type        = string
  description = "constellation kernel cmdline"
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

variable "machine" {
  type        = string
  default     = "q35"
  description = "machine type. use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'"
}

variable "firmware" {
  type        = string
  default     = "/usr/share/OVMF/OVMF_CODE.secboot.fd"
  description = "path to UEFI firmware file. Use \"OVMF_CODE_4M.ms.fd\" on Ubuntu and \"OVMF_CODE.secboot.fd\" on Fedora."
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
