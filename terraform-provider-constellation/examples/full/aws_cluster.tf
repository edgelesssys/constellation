terraform {
  required_providers {
    constellation = {
      source  = "edgelesssys/constellation"
      version = "X.Y.Z"
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
  csp                 = "aws"
  attestation_variant = "aws-sev-snp"
  region              = "us-east-2"
  zone                = "us-east-2c"

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

module "aws_iam" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source      = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/iam/aws"
  name_prefix = "constell"
  region      = local.region
}

module "aws_infrastructure" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/aws"
  name   = "constell"
  node_groups = {
    control_plane_default = {
      role          = "control-plane"
      instance_type = "m6a.xlarge"
      disk_size     = 30
      disk_type     = "gp3"
      initial_count = 2
      zone          = local.zone
    },
    worker_default = {
      role          = "worker"
      instance_type = "m6a.xlarge"
      disk_size     = 30
      disk_type     = "gp3"
      initial_count = 2
      zone          = local.zone
    }
  }
  iam_instance_profile_name_worker_nodes  = module.aws_iam.iam_instance_profile_name_worker_nodes
  iam_instance_profile_name_control_plane = module.aws_iam.iam_instance_profile_name_control_plane
  image_id                                = data.constellation_image.bar.reference
  region                                  = local.region
  zone                                    = local.zone
  debug                                   = false
  enable_snp                              = true
  custom_endpoint                         = ""
}

data "constellation_attestation" "foo" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image_version       = local.version
}

data "constellation_image" "bar" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image_version       = local.version
  region              = local.region
}

resource "constellation_cluster" "aws_example" {
  csp                                = local.csp
  constellation_microservice_version = local.version
  name                               = module.aws_infrastructure.name
  uid                                = module.aws_infrastructure.uid
  image_version                      = local.version
  image_reference                    = data.constellation_image.bar.reference
  attestation                        = data.constellation_attestation.foo.attestation
  init_secret                        = module.aws_infrastructure.init_secret
  master_secret                      = local.master_secret
  master_secret_salt                 = local.master_secret_salt
  measurement_salt                   = local.measurement_salt
  out_of_cluster_endpoint            = module.aws_infrastructure.out_of_cluster_endpoint
  in_cluster_endpoint                = module.aws_infrastructure.in_cluster_endpoint
  network_config = {
    ip_cidr_node    = module.aws_infrastructure.ip_cidr_node
    ip_cidr_service = "10.96.0.0/12"
  }
}

output "kubeconfig" {
  value       = constellation_cluster.aws_example.kubeconfig
  description = "KubeConfig for the Constellation cluster."
}
