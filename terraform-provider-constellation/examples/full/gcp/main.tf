terraform {
  required_providers {
    constellation = {
      source  = "edgelesssys/constellation"
      version = "0.0.0" # replace with the version you want to use
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.0"
    }
  }
}

locals {
  name                = "constell"
  version             = "vX.Y.Z"
  csp                 = "gcp"
  attestation_variant = "gcp-sev-es"
  region              = "europe-west3"
  zone                = "europe-west3-b"
  project_id          = "constellation-331613"

  master_secret      = random_bytes.master_secret.hex
  master_secret_salt = random_bytes.master_secret_salt.hex
  measurement_salt   = random_bytes.measurement_salt.hex
}

resource "random_bytes" "master_secret" {
  length = 32
}

resource "random_bytes" "master_secret_salt" {
  length = 32
}

resource "random_bytes" "measurement_salt" {
  length = 32
}

module "gcp_iam" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source             = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/iam/gcp"
  project_id         = local.project_id
  service_account_id = "${local.name}-test-sa"
  zone               = local.zone
  region             = local.region
}

module "gcp_infrastructure" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/gcp"
  name   = local.name
  node_groups = {
    control_plane_default = {
      role          = "control-plane"
      instance_type = "n2d-standard-4"
      disk_size     = 30
      disk_type     = "pd-ssd"
      initial_count = 2
      zone          = local.zone
    },
    worker_default = {
      role          = "worker"
      instance_type = "n2d-standard-4"
      disk_size     = 30
      disk_type     = "pd-ssd"
      initial_count = 2
      zone          = local.zone
    }
  }
  image_id = data.constellation_image.bar.image.reference
  debug    = false
  zone     = local.zone
  region   = local.region
  project  = local.project_id
}

data "constellation_attestation" "foo" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image               = data.constellation_image.bar.image
}

data "constellation_image" "bar" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  version             = local.version
}

resource "constellation_cluster" "gcp_example" {
  csp                     = local.csp
  name                    = module.gcp_infrastructure.name
  uid                     = module.gcp_infrastructure.uid
  image                   = data.constellation_image.bar.image
  attestation             = data.constellation_attestation.foo.attestation
  init_secret             = module.gcp_infrastructure.init_secret
  master_secret           = local.master_secret
  master_secret_salt      = local.master_secret_salt
  measurement_salt        = local.measurement_salt
  out_of_cluster_endpoint = module.gcp_infrastructure.out_of_cluster_endpoint
  in_cluster_endpoint     = module.gcp_infrastructure.in_cluster_endpoint
  gcp = {
    project_id          = module.gcp_infrastructure.project
    service_account_key = module.gcp_iam.service_account_key
  }
  network_config = {
    ip_cidr_node    = module.gcp_infrastructure.ip_cidr_node
    ip_cidr_service = "10.96.0.0/12"
    ip_cidr_pod     = module.gcp_infrastructure.ip_cidr_pod
  }
}

output "kubeconfig" {
  value       = constellation_cluster.gcp_example.kubeconfig
  sensitive   = true
  description = "KubeConfig for the Constellation cluster."
}
