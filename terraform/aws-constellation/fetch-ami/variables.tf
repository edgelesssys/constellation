variable "attestation_variant" {
  description = "The attestation variant to fetch AMI data for."
  type        = string
}

variable "region" {
  description = "The AWS region to fetch AMI data for."
  type        = string
}

variable "image" {
  description = "The image reference or semantical release version to fetch AMI data for."
  type        = string
}
