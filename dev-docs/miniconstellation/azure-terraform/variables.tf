variable "resource_group" {
  type        = string
  description = "Name of the new resource group for the Miniconstellation cluster"
}

variable "location" {
  type        = string
  description = "The Azure region to create the cluster in (westus|eastus|northeurope|westeurope)"
}

variable "machine_type" {
  type        = string
  description = "The Azure VM type"
  default     = "Standard_D8s_v5"
}
