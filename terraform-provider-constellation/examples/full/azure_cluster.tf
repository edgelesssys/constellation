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
  csp                 = "azure"
  attestation_variant = "azure-sev-snp"
  location            = "northeurope"

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

module "azure_iam" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source                 = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/iam/azure"
  location               = local.location
  service_principal_name = "${local.name}-test-sp"
  resource_group_name    = "${local.name}-test-rg"
}

module "azure_infrastructure" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source                 = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/azure"
  name                   = local.name
  user_assigned_identity = module.azure_iam.uami_id
  node_groups = {
    control_plane_default = {
      role          = "control-plane"
      instance_type = "Standard_DC4as_v5"
      disk_size     = 30
      disk_type     = "Premium_LRS"
      initial_count = 3
    },
    worker_default = {
      role          = "worker"
      instance_type = "Standard_DC4as_v5"
      disk_size     = 30
      disk_type     = "Premium_LRS"
      initial_count = 2
    }
  }
  location       = local.location
  image_id       = data.constellation_image.bar.reference
  resource_group = module.azure_iam.base_resource_group
  create_maa     = true
}

data "constellation_attestation" "foo" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image_version       = local.version
  maa_url             = module.azure_infrastructure.attestation_url
}

data "constellation_image" "bar" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image_version       = local.version
}

resource "constellation_cluster" "azure_example" {
  csp                                = local.csp
  constellation_microservice_version = local.version
  name                               = module.azure_infrastructure.name
  uid                                = module.azure_infrastructure.uid
  image_version                      = local.version
  image_reference                    = data.constellation_image.bar.reference
  attestation                        = data.constellation_attestation.foo.attestation
  init_secret                        = module.azure_infrastructure.init_secret
  master_secret                      = local.master_secret
  master_secret_salt                 = local.master_secret_salt
  measurement_salt                   = local.measurement_salt
  out_of_cluster_endpoint            = module.azure_infrastructure.out_of_cluster_endpoint
  in_cluster_endpoint                = module.azure_infrastructure.in_cluster_endpoint
  azure = {
    tenant_id                   = module.azure_iam.tenant_id
    subscription_id             = module.azure_iam.subscription_id
    uami_client_id              = module.azure_infrastructure.user_assigned_identity_client_id
    uami_resource_id            = module.azure_iam.uami_id
    location                    = local.location
    resource_group              = module.azure_iam.base_resource_group
    load_balancer_name          = module.azure_infrastructure.loadbalancer_name
    network_security_group_name = module.azure_infrastructure.network_security_group_name
  }
  network_config = {
    ip_cidr_node    = module.azure_infrastructure.ip_cidr_node
    ip_cidr_service = "10.96.0.0/12"
  }
}
