variable "name" {
  type        = string
  description = "Base name of the instance group."
}

variable "role" {
  type        = string
  description = "The role of the instance group. Has to be 'ControlPlane' or 'Worker'."
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
