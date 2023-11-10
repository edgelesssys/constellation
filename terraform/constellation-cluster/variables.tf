variable "constellation_version" {
  type        = string
  description = "Constellation CLI version to use."
  default     = "@@CONSTELLATION_VERSION@@"
}

variable "csp" {
  type        = string
  description = "The cloud service provider to use."
  validation {
    condition     = var.csp == "aws" || var.csp == "gcp" || var.csp == "azure"
    error_message = "The cloud service provider to use."
  }
}

variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = optional(number)
    instance_type = string
    disk_size     = number
    disk_type     = string
    zone          = optional(string, "")       # For AWS, GCP
    zones         = optional(list(string), []) # For Azure
  }))
  description = "A map of node group names to node group configurations."
  validation {
    condition     = can([for group in var.node_groups : group.role == "control-plane" || group.role == "worker"])
    error_message = "The role has to be 'control-plane' or 'worker'."
  }
}

variable "name" {
  type        = string
  description = "Name used in the cluster's named resources / cluster name."
}

variable "uid" {
  type        = string
  description = "The UID of the Constellation."
}

variable "clusterEndpoint" {
  type        = string
  description = "Endpoint of the cluster."
}

variable "inClusterEndpoint" {
  type        = string
  description = "The endpoint the cluster uses to reach itself. This might differ from the ClusterEndpoint in case e.g. an internal load balancer is used."
}

variable "initSecretHash" {
  type        = string
  description = "Init secret hash."
}

variable "ipCidrNode" {
  type        = string
  description = "Node IP CIDR."
}

variable "apiServerCertSANs" {
  type        = list(string)
  description = "List of additional SANs (Subject Alternative Names) for the Kubernetes API server certificate."
}

variable "aws_config" {
  type = object({
    region                             = string
    zone                               = string
    iam_instance_profile_worker_nodes  = string
    iam_instance_profile_control_plane = string
  })
  description = "The cluster config for AWS."
  default     = null
}

variable "azure_config" {
  type = object({
    subscription             = string
    tenant                   = string
    location                 = string
    resourceGroup            = string
    userAssignedIdentity     = string
    deployCSIDriver          = bool
    secureBoot               = bool
    maaURL                   = string
    networkSecurityGroupName = string
    loadBalancerName         = string
  })
  description = "The cluster config for Azure."
  default     = null
}

variable "gcp_config" {
  type = object({
    region            = string
    zone              = string
    project           = string
    ipCidrPod         = string
    serviceAccountKey = string
  })
  description = "The cluster config for GCP."
  default     = null
}

variable "image" {
  type        = string
  description = "The node image reference or semantic release version."
  validation {
    condition     = length(var.image) > 0
    error_message = "The image reference must not be empty."
  }
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version."
}

variable "microservice_version" {
  type        = string
  description = "Microservice version."
}

variable "debug" {
  type        = bool
  default     = false
  description = "DON'T USE IN PRODUCTION: Enable debug mode and allow the use of debug images."
}
