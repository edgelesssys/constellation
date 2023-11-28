data "constellation_attestation" "test" {
  csp                 = "aws"
  attestation_variant = "aws-sev-snp"
  image_version       = "v2.13.0"
}
