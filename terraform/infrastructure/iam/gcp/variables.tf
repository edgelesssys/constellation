variable "project_id" {
  type        = string
  description = "ID of the GCP project the cluster should reside in."
}

variable "service_account_id" {
  type        = string
  description = "ID for the service account being created. Must match ^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$."
}

variable "region" {
  type        = string
  description = "GCP region the cluster should reside in. Needs to have the N2D machine type available."
}

variable "zone" {
  type        = string
  description = "GCP zone the cluster should reside in. Needs to be within the specified region."
}
