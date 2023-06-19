variable "name" {
  type        = string
  default     = "constell"
  description = "Base name of the cluster."
}

variable "node_groups" {
  type = map(object({
    role          = string
    zone          = string
    instance_type = string
    disk_size     = number
    disk_type     = string
    initial_count = number
  }))
  description = "A map of node group names to node group configurations."
}

variable "project" {
  type        = string
  description = "The GCP project to deploy the cluster in."
}

variable "region" {
  type        = string
  description = "The GCP region to deploy the cluster in."
}

variable "zone" {
  type        = string
  description = "The GCP zone to deploy the cluster in."
}

variable "image_id" {
  type        = string
  description = "The GCP image to use for the cluster nodes."
}

variable "debug" {
  type        = bool
  default     = false
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
}
