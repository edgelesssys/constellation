terraform {
  required_providers {
    constellation = {
      source = "registry.terraform.io/edgelesssys/constellation"
    }
  }
}

provider "constellation" {
}

data "constellation_attestation" "test" {
  csp                 = "aws"
  attestation_variant = "aws-sev-snp"
  image_version       = "v2.13.0"
}

output "measurements" {
  value = data.constellation_attestation.test.measurements
}

output "attestation" {
  value = data.constellation_attestation.test.attestation
}
