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

variable "credentials_file" {
  type        = string
  description = "The path to the GCP credentials file."
}

variable "instance_type" {
  type        = string
  description = "The GCP instance type to deploy."
}

variable "state_disk_type" {
  type        = string
  default     = "pd-ssd"
  description = "The type of the state disk."
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
