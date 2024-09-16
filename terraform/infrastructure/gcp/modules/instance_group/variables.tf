variable "base_name" {
  type        = string
  description = "Base name of the instance group."
}

variable "node_group_name" {
  type        = string
  description = "Constellation name for the node group (used for configuration and CSP-independent naming)."
}

variable "role" {
  type        = string
  description = "Role of the instance group."
  validation {
    condition     = contains(["control-plane", "worker"], var.role)
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "uid" {
  type        = string
  description = "Unique Identifier (UID) of the cluster."
}

variable "labels" {
  type        = map(string)
  default     = {}
  description = "Labels to apply to the instance group."
}

variable "instance_type" {
  type        = string
  description = "Instance type for the nodes."
}

variable "initial_count" {
  type        = number
  description = "Number of instances in the group."
}

variable "image_id" {
  type        = string
  description = "OS Image reference for the cluster's nodes."
}

variable "disk_size" {
  type        = number
  description = "Disk size for the state disk of the nodes [GB]."
}

variable "disk_type" {
  type        = string
  description = "Disk type for the nodes. Has to be either 'pd-standard' or 'pd-ssd'."
}

variable "network" {
  type        = string
  description = "Name of the network to use."
}

variable "subnetwork" {
  type        = string
  description = "Name of the subnetwork to use."
}

variable "kube_env" {
  type        = string
  description = "Value of the \"kube-env\" metadata key."
}

variable "init_secret_hash" {
  type        = string
  description = "BCrypt Hash of the initialization secret."
}

variable "named_ports" {
  type        = list(object({ name = string, port = number }))
  default     = []
  description = "Named ports for the instance group."
}

variable "debug" {
  type        = bool
  default     = false
  description = "DO NOT USE IN PRODUCTION. Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}

variable "console_access" {
  type        = bool
  description = "Enable serial console access to OS images that expose a serial console. This will be shadowed by `debug` (i.e. if `debug` is enabled, console access will be enabled)."
}

variable "alias_ip_range_name" {
  type        = string
  description = "Name of the alias IP range to use."
}

variable "zone" {
  type        = string
  description = "Zone to deploy the instance group in."
}

variable "custom_endpoint" {
  type        = string
  description = "Custom endpoint to use for the Kubernetes API server. If not set, the default endpoint will be used."
}

variable "cc_technology" {
  type        = string
  description = "The confidential computing technology to use for the nodes. One of `SEV`, `SEV_SNP`."
  validation {
    condition     = contains(["SEV", "SEV_SNP"], var.cc_technology)
    error_message = "The confidential computing technology has to be 'SEV' or 'SEV_SNP'."
  }
}
