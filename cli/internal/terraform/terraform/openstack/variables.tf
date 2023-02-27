variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "control_plane_count" {
  type        = number
  description = "The number of control plane nodes to deploy."
}

variable "worker_count" {
  type        = number
  description = "The number of worker nodes to deploy."
}

variable "state_disk_size" {
  type        = number
  default     = 30
  description = "The size of the state disk in GB."
}

variable "availability_zone" {
  type        = string
  description = "The availability zone to deploy the nodes in."
}

variable "image_url" {
  type        = string
  description = "The image to use for cluster nodes."
}

variable "direct_download" {
  type        = bool
  description = "If enabled, downloads OS image directly from source URL to OpenStack. Otherwise, downloads image to local machine and uploads to OpenStack."
}

variable "flavor_id" {
  type        = string
  description = "The flavor (machine type) to use for cluster nodes."
}

variable "floating_ip_pool_id" {
  type        = string
  description = "The pool (network name) to use for floating IPs."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}
