variable "csp" {
  description = "The cloud service provider to fetch image data for."
  type        = string
}

variable "attestation_variant" {
  description = "The attestation variant to fetch image data for."
  type        = string
}

variable "region" {
  description = "The region to fetch image data for."
  type        = string
  default     = ""
}

variable "image" {
  description = "The image reference or semantical release version to fetch image data for."
  type        = string
}
