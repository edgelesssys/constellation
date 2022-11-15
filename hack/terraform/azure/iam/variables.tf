variable "resource_group_name" {
  type        = string
  description = "Resource group name"
  default     = "tf_iam_test_rg"
}

variable "service_principal_name" {
  type        = string
  description = "Service principal name"
  default     = "tf_iam_test_principal"
}

variable "region" {
  type        = string
  description = "Azure resource location. Defaults to westus"
  default     = "westus"
}
