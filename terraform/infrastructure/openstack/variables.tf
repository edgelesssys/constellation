variable "node_groups" {
  type = map(object({
    role            = string
    initial_count   = number // number of instances in the node group
    flavor_id       = string // flavor (machine type) to use for instances
    state_disk_size = number // size of state disk (GiB)
    state_disk_type = string // type of state disk. Can be 'standard' or 'premium'
    zone            = string // availability zone
  }))

  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }

  description = "A map of node group names to node group configurations."
}

variable "cloud" {
  type        = string
  default     = null
  description = "The cloud to use within the OpenStack \"clouds.yaml\" file. Optional. If not set, environment variables are used."
}

variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "image_url" {
  type        = string
  description = "The image to use for cluster nodes."
}

variable "direct_download" {
  type        = bool
  description = "If enabled, downloads OS image directly from source URL to OpenStack. Otherwise, downloads image to local machine and uploads to OpenStack."
}

variable "floating_ip_pool_id" {
  type        = string
  description = "The pool (network name) to use for floating IPs."
}

variable "openstack_user_domain_name" {
  type        = string
  description = "OpenStack user domain name."
}

variable "openstack_username" {
  type        = string
  description = "OpenStack user name."
}

variable "openstack_password" {
  type        = string
  description = "OpenStack password."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes apiserver. If not set, the default endpoint will be used."
}
