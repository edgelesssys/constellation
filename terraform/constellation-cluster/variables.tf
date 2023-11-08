variable "csp" {
  type        = string
  description = "The cloud service provider to use."
  validation {
    condition     = var.csp == "aws" || var.csp == "gcp"
    error_message = "The CSP must be one of {aws|gcp}."
  }
}

variable "node_groups" {
  type = map(object({
    role          = string
    initial_count = optional(number)
    instance_type = string
    disk_size     = number
    disk_type     = string
    zone          = string
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

variable "gcp_config" {
  type = object({
    region                = string
    zone                  = string
    project               = string
    serviceAccountKeyPath = string
    ipCidrPod             = string
  })
  description = "The cluster config for GCP."
  default     = null
}

variable "image" {
  type        = string
  description = "The node image reference or semantical release version."
  validation {
    condition     = length(var.image) > 0
    error_message = "The image variable must not be empty."
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
