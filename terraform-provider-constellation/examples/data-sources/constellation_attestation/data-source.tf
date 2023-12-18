data "constellation_image" "example" {} # Fill accordingly for the CSP

data "constellation_attestation" "test" {
  csp                 = "aws"
  attestation_variant = "aws-sev-snp"
  image               = data.constellation_image.example.image
}
