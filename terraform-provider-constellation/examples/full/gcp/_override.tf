terraform {
  required_providers {
    constellation = {
      source  = "edgelesssys/constellation"
      version = "2.15.0-pre.0.20231221113030-1a5e64ef55c1"
    }
  }
}
locals {
  name    = "e2e-521-1" # try -1
  version = "ref/main/stream/nightly/v2.14.0-pre.0.20231214193540-2c50abcc919b"
}
module "gcp_iam" {
  #  source = "../terraform-module/iam/gcp"
  source = "../../../../terraform/infrastructure/iam/gcp"
}
module "gcp_infrastructure" {
  #  source = "../terraform-module/gcp"
  source = "../../../../terraform/infrastructure/gcp"
}
locals {
  region = "europe-west3"
  zone   = "europe-west3-b"
}
