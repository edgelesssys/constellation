terraform {
  required_providers {
    constellation = {
      source  = "edgelesssys/constellation"
      version = "2.15.0-pre.0.20231222082729-7992d719d78c"
    }
  }
}
locals {
  name    = "e2e-107"
  version = "ref/main/stream/nightly/v2.14.0-pre.0.20231214193540-2c50abcc919b"
}
module "azure_iam" {
  #  source = "terraform-module/iam/azure"
  source = "../../../../terraform/infrastructure/iam/azure"
}

module "azure_infrastructure" {
  #  source = "../terraform-module/azure"
  source = "../../../../terraform/infrastructure/azure"
}

#locals {
#  version            = "v2.14.0"
#}
