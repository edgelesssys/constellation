variable "project_id" {
  type        = string
  description = "GCP Project ID."
}

variable "service_account_id" {
  type        = string
  description = "ID for the service account being created. Must match ^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$."
}

variable "region" {
  type        = string
  description = "Region used for constellation clusters. Needs to have the N2D machine type available."
}

variable "zone" {
  type        = string
  description = "Zone used for constellation clusters. Needs to be within the specified region."
}
