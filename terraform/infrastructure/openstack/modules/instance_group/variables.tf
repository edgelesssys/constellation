variable "base_name" {
  type        = string
  description = "Base name of the instance group."
}

variable "node_group_name" {
  type        = string
  description = "Constellation name for the instance group (used for configuration and CSP-independent naming)."
}

variable "role" {
  type        = string
  description = "The role of the instance group."
  validation {
    condition     = contains(["control-plane", "worker"], var.role)
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "tags" {
  type        = list(string)
  description = "Tags to attach to each node."
}

variable "uid" {
  type        = string
  description = "Unique ID of the Constellation."
}

variable "initial_count" {
  type        = number
  description = "Number of instances in this instance group."
}

variable "image_id" {
  type        = string
  description = "OS Image reference for the cluster's nodes."
}

variable "flavor_id" {
  type        = string
  description = "Flavor ID (machine type) to use for the nodes."
}

variable "security_groups" {
  type        = list(string)
  description = "Security groups to place the nodes in."
}

variable "disk_size" {
  type        = number
  description = "Disk size for the state disk of the nodes [GB]."
}

variable "state_disk_type" {
  type        = string
  description = "Type of the state disk."
}

variable "availability_zone" {
  type        = string
  description = "Availability zone to deploy the nodes in."
}

variable "network_id" {
  type        = string
  description = "Network ID to attach each node to."
}

variable "subnet_id" {
  type        = string
  description = "Subnetwork ID to attach each node to."
}

variable "init_secret_hash" {
  type        = string
  description = "Hash of the init secret."
}

variable "identity_internal_url" {
  type        = string
  description = "Internal URL of the Identity service."
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

variable "openstack_region_name" {
  type        = string
  description = "OpenStack region name."
}

variable "openstack_load_balancer_endpoint" {
  type        = string
  description = "OpenStack load balancer endpoint."
}
