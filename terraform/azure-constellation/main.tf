module "azure_iam" {
  source                 = "../infrastructure/iam/azure"
  region                 = var.location
  service_principal_name = var.service_principal_name
  resource_group_name    = var.resource_group_name
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
  csp                 = "azure"
  attestation_variant = "azure-sev-snp"
  image               = var.image
  depends_on          = [null_resource.ensure_yq]
}

module "azure" {
  source                 = "../infrastructure/azure"
  name                   = var.name
  user_assigned_identity = module.azure_iam.uami_id
  node_groups            = var.node_groups
  location               = var.location
  image_id               = module.fetch_image.image
  debug                  = var.debug
  resource_group         = module.azure_iam.base_resource_group
  create_maa             = var.create_maa
}

module "constellation" {
  source               = "../constellation-cluster"
  csp                  = "azure"
  debug                = var.debug
  name                 = var.name
  image                = var.image
  microservice_version = var.microservice_version
  kubernetes_version   = var.kubernetes_version
  uid                  = module.azure.uid
  clusterEndpoint      = module.azure.out_of_cluster_endpoint
  inClusterEndpoint    = module.azure.in_cluster_endpoint
  initSecretHash       = module.azure.initSecret
  ipCidrNode           = module.azure.ip_cidr_nodes
  apiServerCertSANs    = module.azure.api_server_cert_sans
  node_groups          = var.node_groups
  azure_config = {
    subscription             = module.azure_iam.subscription_id
    tenant                   = module.azure_iam.tenant_id
    location                 = var.location
    resourceGroup            = module.azure.resource_group
    userAssignedIdentity     = module.azure_iam.uami_id
    deployCSIDriver          = var.deploy_csi_driver
    secureBoot               = var.secure_boot
    maaURL                   = module.azure.attestationURL
    networkSecurityGroupName = module.azure.network_security_group_name
    loadBalancerName         = module.azure.loadbalancer_name
  }
  depends_on = [module.azure, module.azure_iam, null_resource.ensure_yq]
}
