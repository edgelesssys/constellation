variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
}

variable "image" {
  type        = string
  description = "Node image reference or semantic release version. When not set, the latest default version will be used."
  default     = ""
}

variable "microservice_version" {
  type        = string
  description = "Microservice version. When not set, the latest default version will be used."
  default     = ""
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version. When not set, the latest default version will be used."
  default     = ""
}

variable "debug" {
  type        = bool
  default     = false
  description = "DON'T USE IN PRODUCTION: Enable debug mode and allow the use of debug images."
}

variable "custom_endpoint" {
  type        = string
  default     = ""
  description = "Custom endpoint (DNS Name) to use for the Constellation API server. If not set, the default endpoint will be used."
}

variable "internal_load_balancer" {
  type        = bool
  default     = false
  description = "Use an internal load balancer."
}

variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = optional(number)
    instance_type = string
    disk_size     = number
    disk_type     = string
    zones         = optional(list(string))
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "service_principal_name" {
  type        = string
  description = "Name of the service principal used to create the cluster."
}

variable "resource_group_name" {
  type        = string
  description = "Name of the resource group the cluster's resources are created in."
}

variable "location" {
  type        = string
  description = "Azure datacenter region the cluster will be deployed in."
}

variable "deploy_csi_driver" {
  type        = bool
  default     = true
  description = "Deploy the Azure Disk CSI driver with on-node encryption into the cluster."
}

variable "secure_boot" {
  type        = bool
  default     = false
  description = "Enable secure boot for VMs. If enabled, the OS image has to include a virtual machine guest state (VMGS) blob."
}

variable "create_maa" {
  type        = bool
  default     = true
  description = "Create an MAA for attestation."
}
