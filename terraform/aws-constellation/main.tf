locals {
  region = substr(var.zone, 0, length(var.zone) - 1)
}

module "aws_iam" {
  source      = "../infrastructure/iam/aws"
  name_prefix = var.name_prefix
  region      = local.region
}


resource "null_resource" "ensure_yq" {
  provisioner "local-exec" {
    command = <<EOT
         ../common/install-yq.sh
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
}

module "fetch_image" {
  source              = "../common/fetch-image"
  csp                 = "aws"
  attestation_variant = "aws-sev-snp"
  region              = local.region
  image               = var.image
  depends_on          = [module.aws_iam, null_resource.ensure_yq]
}


module "aws" {
  source                             = "../infrastructure/aws"
  name                               = var.name
  node_groups                        = var.node_groups
  iam_instance_profile_worker_nodes  = module.aws_iam.worker_nodes_instance_profile
  iam_instance_profile_control_plane = module.aws_iam.control_plane_instance_profile
  ami                                = module.fetch_image.image
  region                             = local.region
  zone                               = var.zone
  debug                              = var.debug
  enable_snp                         = var.enable_snp
  custom_endpoint                    = var.custom_endpoint
}

module "constellation" {
  source               = "../constellation-cluster"
  csp                  = "aws"
  name                 = var.name
  image                = var.image
  microservice_version = var.microservice_version
  kubernetes_version   = var.kubernetes_version
  uid                  = module.aws.uid
  clusterEndpoint      = module.aws.out_of_cluster_endpoint
  inClusterEndpoint    = module.aws.in_cluster_endpoint
  initSecretHash       = module.aws.initSecret
  ipCidrNode           = module.aws.ip_cidr_nodes
  apiServerCertSANs    = module.aws.api_server_cert_sans
  node_groups          = var.node_groups
  aws_config = {
    region                             = local.region
    zone                               = var.zone
    iam_instance_profile_worker_nodes  = module.aws_iam.worker_nodes_instance_profile
    iam_instance_profile_control_plane = module.aws_iam.control_plane_instance_profile
  }
  depends_on = [module.aws, null_resource.ensure_yq]
}
