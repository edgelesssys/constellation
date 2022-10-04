variable "name" {
  type        = string
  description = "Name of your Constellation"
}

variable "iam_instance_profile_worker_nodes" {
  type        = string
  description = "Name of the IAM instance profile for worker nodes"
}

variable "iam_instance_profile_control_plane" {
  type        = string
  description = "Name of the IAM instance profile for control plane nodes"
}

variable "instance_type" {
  type        = string
  description = "Instance type for worker nodes"
  default     = "t3.micro"
}

variable "state_disk_size" {
  type        = number
  description = "Disk size for nodes [GB]"
  default     = 30
}

variable "control_plane_count" {
  type        = number
  description = "Number of control plane nodes"
  default     = 1
}

variable "worker_count" {
  type        = number
  description = "Number of worker nodes"
  default     = 1
}

variable "ami" {
  type        = string
  description = "AMI ID"
  default     = "ami-02f3416038bdb17fb" // Ubuntu 22.04 LTS
}

variable "region" {
  type        = string
  description = "The AWS region to create the cluster in"
  default     = "us-east-2"
}

variable "zone" {
  type        = string
  description = "The AWS availability zone name to create the cluster in"
  default     = "us-east-2a"
}

variable "debug" {
  type        = bool
  description = "Enable debug mode. This opens up a debugd port that can be used to deploy a custom bootstrapper."
  default     = false
}
