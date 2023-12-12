locals {
  region = substr(var.zone, 0, length(var.zone) - 2)
}

module "gcp_iam" {
  source             = "../../infrastructure/iam/gcp"
  project_id         = var.project
  service_account_id = var.service_account_id
  region             = local.region
  zone               = var.zone
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
  csp                 = "gcp"
  attestation_variant = "gcp-sev-es"
  image               = var.image
  depends_on          = [null_resource.ensure_yq]
}


module "gcp" {
  source          = "../../infrastructure/gcp"
  project         = var.project
  image_id        = module.fetch_image.image
  name            = var.name
  node_groups     = var.node_groups
  region          = local.region
  zone            = var.zone
  debug           = var.debug
  custom_endpoint = var.custom_endpoint
}

module "constellation" {
  source               = "../constellation-cluster"
  csp                  = "gcp"
  debug                = var.debug
  name                 = var.name
  image                = var.image
  microservice_version = var.microservice_version
  kubernetes_version   = var.kubernetes_version
  uid                  = module.gcp.uid
  clusterEndpoint      = module.gcp.out_of_cluster_endpoint
  inClusterEndpoint    = module.gcp.in_cluster_endpoint
  initSecretHash       = module.gcp.init_secret
  ipCidrNode           = module.gcp.ip_cidr_nodes
  apiServerCertSANs    = module.gcp.extra_api_server_cert_sans
  node_groups          = var.node_groups
  gcp_config = {
    region            = local.region
    zone              = var.zone
    project           = var.project
    ipCidrPod         = module.gcp.ip_cidr_pods
    serviceAccountKey = module.gcp_iam.sa_key
  }
  depends_on = [module.gcp, null_resource.ensure_yq]
}
