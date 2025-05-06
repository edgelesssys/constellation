terraform {
  required_providers {
    constellation = {
      source  = "edgelesssys/constellation"
      version = "2.23.0" // replace with the version you want to use
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

locals {
  name                 = "constell"
  image_version        = "vX.Y.Z"
  kubernetes_version   = "vX.Y.Z"
  microservice_version = "vX.Y.Z"
  csp                  = "azure"
  attestation_variant  = "azure-sev-snp"
  location             = "northeurope"
  control_plane_count  = 3
  worker_count         = 2
  instance_type        = "Standard_DC4as_v5" // Adjust if using TDX
  subscription_id      = "00000000-0000-0000-0000-000000000000"

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
  subscription_id        = local.subscription_id
  location               = local.location
  service_principal_name = "${local.name}-sp"
  resource_group_name    = "${local.name}-rg"
}

module "azure_infrastructure" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source                 = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/azure"
  subscription_id        = local.subscription_id
  name                   = local.name
  user_assigned_identity = module.azure_iam.uami_id
  node_groups = {
    control_plane_default = {
      role          = "control-plane"
      instance_type = local.instance_type
      disk_size     = 30
      disk_type     = "Premium_LRS"
      initial_count = local.control_plane_count
    },
    worker_default = {
      role          = "worker"
      instance_type = local.instance_type
      disk_size     = 30
      disk_type     = "Premium_LRS"
      initial_count = local.worker_count
    }
  }
  location               = local.location
  image_id               = data.constellation_image.bar.image.reference
  resource_group         = module.azure_iam.base_resource_group
  internal_load_balancer = false
  create_maa             = true
}

data "constellation_attestation" "foo" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image               = data.constellation_image.bar.image
  # Needs to be patched manually, see:
  # https://docs.edgeless.systems/constellation/workflows/terraform-provider#quick-setup
  maa_url = module.azure_infrastructure.attestation_url
}

data "constellation_image" "bar" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  version             = local.image_version
}

resource "constellation_cluster" "azure_example" {
  csp                                = local.csp
  name                               = module.azure_infrastructure.name
  uid                                = module.azure_infrastructure.uid
  image                              = data.constellation_image.bar.image
  attestation                        = data.constellation_attestation.foo.attestation
  kubernetes_version                 = local.kubernetes_version
  constellation_microservice_version = local.microservice_version
  init_secret                        = module.azure_infrastructure.init_secret
  master_secret                      = local.master_secret
  master_secret_salt                 = local.master_secret_salt
  measurement_salt                   = local.measurement_salt
  out_of_cluster_endpoint            = module.azure_infrastructure.out_of_cluster_endpoint
  in_cluster_endpoint                = module.azure_infrastructure.in_cluster_endpoint
  api_server_cert_sans               = module.azure_infrastructure.api_server_cert_sans
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

output "maa_url" {
  value       = module.azure_infrastructure.attestation_url
  description = "URL of the MAA provider, required for manual patching."
}

output "kubeconfig" {
  value       = constellation_cluster.azure_example.kubeconfig
  sensitive   = true
  description = "KubeConfig for the Constellation cluster."
}
