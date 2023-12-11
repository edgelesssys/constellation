variable "name_prefix" {
  type        = string
  description = "Prefix for all resources to easily identify related IAM resources."
}

variable "region" {
  type        = string
  description = "AWS region."
  default     = "us-east-2"
}
