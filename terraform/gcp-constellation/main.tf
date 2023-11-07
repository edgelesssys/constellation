locals {
  region = substr(var.zone, 0, length(var.zone) - 2)
}

module "gcp_iam" {
  source             = "../infrastructure/iam/gcp"
  project_id         = var.project
  service_account_id = var.service_account_id
  region             = local.region
  zone               = var.zone
}


resource "null_resource" "ensure_yq" {
  provisioner "local-exec" {
    command = <<EOT
         ${path.module}/install-yq.sh
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
}

module "fetch_image" {
  source              = "../fetch-image"
  csp                 = "gcp"
  attestation_variant = "gcp-sev-es"
  image               = var.image
  depends_on          = [null_resource.ensure_yq]
}


module "gcp" {
  source          = "../infrastructure/gcp"
  project         = var.project
  image_id        = module.fetch_image.image
  name            = var.name
  node_groups     = var.node_groups
  region          = local.region
  zone            = var.zone
  debug           = var.debug
  custom_endpoint = var.custom_endpoint
}

resource "null_resource" "sa_account_file" {
  provisioner "local-exec" {
    command = <<EOT
          #echo "${module.gcp_iam.sa_key}" TODO use base64decode fn
          echo ${module.gcp_iam.sa_key} | base64 -d > "sa_account_file.json"

    EOT
  }
  provisioner "local-exec" {
    when    = destroy
    command = "rm sa_account_file.json"
  }
  triggers = {
    always_run = timestamp()
  }
}


module "constellation" {
  source               = "../constellation-cluster"
  csp                  = "gcp"
  name                 = var.name
  image                = var.image
  microservice_version = var.microservice_version
  kubernetes_version   = var.kubernetes_version
  uid                  = module.gcp.uid
  clusterEndpoint      = module.gcp.out_of_cluster_endpoint
  inClusterEndpoint    = module.gcp.in_cluster_endpoint
  initSecretHash       = module.gcp.initSecret
  ipCidrNode           = module.gcp.ip_cidr_nodes
  apiServerCertSANs    = module.gcp.api_server_cert_sans
  node_groups          = var.node_groups
  gcp_config = {
    region                = local.region
    zone                  = var.zone
    serviceAccountKeyPath = "sa_account_file.json"
    project               = var.project
    ipCidrPod             = module.gcp.ip_cidr_pods
  }
  depends_on = [module.gcp, null_resource.sa_account_file, null_resource.ensure_yq]
}
