variable "name_prefix" {
  type        = string
  description = "Prefix for all resources"
}

variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-2"
}
