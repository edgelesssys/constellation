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
  description = "The role of the instance group."
  validation {
    condition     = contains(["ControlPlane", "Worker"], var.role)
    error_message = "The role has to be 'ControlPlane' or 'Worker'."
  }
}

variable "uid" {
  type        = string
  description = "UID of the cluster. This is used for tags."
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

variable "instance_count" {
  type        = number
  description = "Number of instances in the instance group."
}

variable "image_id" {
  type        = string
  description = "Image ID for the nodes."
}

variable "disk_size" {
  type        = number
  description = "Disk size for the nodes, in GB."
}

variable "disk_type" {
  type        = string
  description = "Disk type for the nodes. Has to be 'pd-standard' or 'pd-ssd'."
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
  description = "Kubernetes env."
}

variable "init_secret_hash" {
  type        = string
  description = "Hash of the init secret."
}

variable "named_ports" {
  type        = list(object({ name = string, port = number }))
  default     = []
  description = "Named ports for the instance group."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This will enable serial port access on the instances."
}

variable "alias_ip_range_name" {
  type        = string
  description = "Name of the alias IP range to use."
}

variable "zone" {
  type        = string
  description = "Zone to deploy the instance group in."
}
