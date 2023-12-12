variable "resource_group_name" {
  type        = string
  description = "Name for the resource group the cluster should reside in."
}

variable "service_principal_name" {
  type        = string
  description = "Name for the service principal used within the cluster."
}

variable "region" {
  type        = string
  description = "Azure region the cluster should reside in."
}
