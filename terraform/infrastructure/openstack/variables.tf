# Variables common to all CSPs

variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}
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

variable "image_id" {
  type        = string
  description = "OS image ID for the cluster's nodes."
}

variable "debug" {
  type        = bool
  default     = false
  description = "DO NOT USE IN PRODUCTION. Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint to use for the Kubernetes API server. If not set, the default endpoint will be used."
}

# OpenStack-specific variables

variable "cloud" {
  type        = string
  default     = null
  description = "Cloud to use within the OpenStack \"clouds.yaml\" file. Optional. If not set, environment variables are used."
}

variable "openstack_clouds_yaml_path" {
  type        = string
  default     = "~/.config/openstack/clouds.yaml"
  description = "Path to OpenStack clouds.yaml file"
}

variable "floating_ip_pool_id" {
  type        = string
  description = "Pool (network name) to use for floating IPs."
}

variable "additional_tags" {
  type        = list(any)
  default     = []
  description = "Additional tags that should be applied to created resources."
}

# STACKIT-specific variables

variable "stackit_project_id" {
  type        = string
  description = "STACKIT project ID."
}

variable "emergency_ssh" {
  type        = bool
  default     = false
  description = "Wether to expose the SSH port through the public load balancer."
}
