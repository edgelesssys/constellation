variable "name_prefix" {
  type        = string
  description = "Name prefix to use on named resources."
}

variable "region" {
  type        = string
  description = "AWS region."
  default     = "us-east-2"
}
