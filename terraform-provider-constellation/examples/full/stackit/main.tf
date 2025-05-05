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
  name                       = "constell"
  image_version              = "vX.Y.Z"
  kubernetes_version         = "vX.Y.Z"
  microservice_version       = "vX.Y.Z"
  csp                        = "stackit"
  attestation_variant        = "qemu-vtpm"
  zone                       = "eu01-1"
  cloud                      = "stackit"
  clouds_yaml_path           = "~/.config/openstack/clouds.yaml"
  floating_ip_pool_id        = "970ace5c-458f-484a-a660-0903bcfd91ad"
  stackit_project_id         = "" // replace with the STACKIT project id
  control_plane_count        = 3
  worker_count               = 2
  instance_type              = "m1a.8cd"
  deploy_yawol_load_balancer = true
  yawol_image_id             = "bcd6c13e-75d1-4c3f-bf0f-8f83580cc1be"
  yawol_flavor_id            = "3b11b27e-6c73-470d-b595-1d85b95a8cdf"

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

module "stackit_infrastructure" {
  // replace $VERSION with the Constellation version you want to use, e.g., v2.14.0
  source = "https://github.com/edgelesssys/constellation/releases/download/$VERSION/terraform-module.zip//terraform-module/openstack"
  name   = local.name
  node_groups = {
    control_plane_default = {
      role            = "control-plane"
      flavor_id       = local.instance_type
      state_disk_size = 30
      state_disk_type = "storage_premium_perf6"
      initial_count   = local.control_plane_count
      zone            = local.zone
    },
    worker_default = {
      role            = "worker"
      flavor_id       = local.instance_type
      state_disk_size = 30
      state_disk_type = "storage_premium_perf6"
      initial_count   = local.worker_count
      zone            = local.zone
    }
  }
  image_id                   = data.constellation_image.bar.image.reference
  debug                      = false
  cloud                      = local.cloud
  openstack_clouds_yaml_path = local.clouds_yaml_path
  floating_ip_pool_id        = local.floating_ip_pool_id
  stackit_project_id         = local.stackit_project_id
}

data "constellation_attestation" "foo" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  image               = data.constellation_image.bar.image
}

data "constellation_image" "bar" {
  csp                 = local.csp
  attestation_variant = local.attestation_variant
  version             = local.image_version
  marketplace_image   = true
}

resource "constellation_cluster" "stackit_example" {
  csp                                = local.csp
  name                               = module.stackit_infrastructure.name
  uid                                = module.stackit_infrastructure.uid
  image                              = data.constellation_image.bar.image
  attestation                        = data.constellation_attestation.foo.attestation
  kubernetes_version                 = local.kubernetes_version
  constellation_microservice_version = local.microservice_version
  init_secret                        = module.stackit_infrastructure.init_secret
  master_secret                      = local.master_secret
  master_secret_salt                 = local.master_secret_salt
  measurement_salt                   = local.measurement_salt
  out_of_cluster_endpoint            = module.stackit_infrastructure.out_of_cluster_endpoint
  in_cluster_endpoint                = module.stackit_infrastructure.in_cluster_endpoint
  api_server_cert_sans               = module.stackit_infrastructure.api_server_cert_sans
  openstack = {
    cloud                      = local.cloud
    clouds_yaml_path           = local.clouds_yaml_path
    floating_ip_pool_id        = local.floating_ip_pool_id
    deploy_yawol_load_balancer = local.deploy_yawol_load_balancer
    yawol_image_id             = local.yawol_image_id
    yawol_flavor_id            = local.yawol_flavor_id
    network_id                 = module.stackit_infrastructure.network_id
    subnet_id                  = module.stackit_infrastructure.lb_subnetwork_id
  }
  network_config = {
    ip_cidr_node    = module.stackit_infrastructure.ip_cidr_node
    ip_cidr_service = "10.96.0.0/12"
  }
}

output "kubeconfig" {
  value       = constellation_cluster.stackit_example.kubeconfig
  sensitive   = true
  description = "KubeConfig for the Constellation cluster."
}
