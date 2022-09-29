variable "name" {
  type        = string
  description = "Name of your Constellation"
}

variable "worker_nodes_iam_instance_profile" {
  type        = string
  description = "Name of the IAM instance profile for worker nodes"
}

variable "control_plane_iam_instance_profile" {
  type        = string
  description = "Name of the IAM instance profile for control plane nodes"
}

variable "instance_type" {
  type        = string
  description = "Instance type for worker nodes"
  default     = "t2.micro"
}

variable "disk_size" {
  type        = number
  description = "Disk size for nodes [GB]"
  default     = 30
}

variable "count_control_plane" {
  type        = number
  description = "Number of control plane nodes"
  default     = 1
}

variable "count_worker_nodes" {
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
  description = "AWS region"
  default     = "us-east-2"
}
