terraform {
  required_providers {
    constellation = {
      source = "registry.terraform.io/edgelesssys/constellation"
    }
  }
}

provider "constellation" {
  example_value = "test"
}
