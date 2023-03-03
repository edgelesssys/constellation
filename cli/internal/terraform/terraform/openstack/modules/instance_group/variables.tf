variable "name" {
  type        = string
  description = "Base name of the instance group."
}

variable "uid" {
  type        = string
  description = "Unique ID of the Constellation."
}

variable "role" {
  type        = string
  description = "The role of the instance group."
  validation {
    condition     = contains(["ControlPlane", "Worker"], var.role)
    error_message = "The role has to be 'ControlPlane' or 'Worker'."
  }
}

variable "instance_count" {
  type        = number
  description = "Number of instances in the instance group."
}

variable "image_id" {
  type        = string
  description = "Image ID for the nodes."
}

variable "flavor_id" {
  type        = string
  description = "Flavor ID (machine type) to use for the nodes."
}

variable "security_groups" {
  type        = list(string)
  description = "Security groups to place the nodes in."
}

variable "tags" {
  type        = list(string)
  description = "Tags to attach to each node."
}

variable "disk_size" {
  type        = number
  description = "Disk size for the nodes, in GiB."
}

variable "availability_zone" {
  type        = string
  description = "The availability zone to deploy the nodes in."
}

variable "network_id" {
  type        = string
  description = "Network ID to attach each node to."
}

variable "init_secret_hash" {
  type        = string
  description = "Hash of the init secret."
}

variable "identity_internal_url" {
  type        = string
  description = "Internal URL of the Identity service."
}


variable "openstack_service_account_token" {
  type        = string
  description = "OpenStack service account token."
}
