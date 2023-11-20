variable "resource_group_location" {
  type        = string
  default     = "westeurope"
  description = "Location of the resource group."
}

variable "name_prefix" {
  type    = string
  default = "vpn-test"
}

variable "remote_addr" {
  type        = string
  description = "The public IP address of the remote host."
}

variable "ike_psk" {
  type        = string
  description = "The IKE pre-shared key."
}

variable "local_ts" {
  type        = string
  description = "The local traffic selector."
  default     = "10.99.0.0/16"
}

variable "remote_ts" {
  type        = list(string)
  description = "The remote traffic selector."
  default     = ["10.10.0.0/16", "10.96.0.0/12"]
}
