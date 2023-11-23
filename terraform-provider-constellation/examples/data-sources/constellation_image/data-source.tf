data "constellation_image" "example" {
  image_version       = "v2.13.0"
  attestation_variant = "aws-sev-snp"
  csp                 = "aws"
  region              = "eu-west-1"
}
