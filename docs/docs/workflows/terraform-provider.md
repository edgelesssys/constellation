# Use the Terraform provider
You can manage a Constellation cluster through Terraform.
<!-- TODO(elchead): check link during release -->
The provider is available through the [Terraform registry](https://registry.terraform.io/providers/edgelesssys/constellation/latest) and is released as part of Constellation releases. This page shows how to use the Constellation Terraform provider.

## Prerequisites
- a Linux / Mac operating system (ARM64/AMD64)
- a Terraform installation of version `v1.4.4` or above

## Quick setup
The example shows how to set up a Constellation cluster with the default infrastructure and IAM setup. The modules can either be consumed by using the remote source shown below (recommended) or by downloading them from the [Constellation release page ](https://github.com/edgelesssys/constellation/releases/latest) and placing them in the Terraform workspace directory.


1. Create a directory (workspace) for your Constellation cluster.
  ```bash
  mkdir constellation-workspace
  cd constellation-workspace
  ```

1. Create a `main.tf` file.
<!-- TODO: put file in repo to reuse in e2e test? -->

  <tabs groupId="csp">

  <tabItem value="azure" label="Azure">

  ```
  terraform {
  required_providers {
    constellation = {
      source  = "tbd/constellation"
      version = "2.13.0"
    }
  }
}

data "constellation_attestation" "att" {
    csp = "azure"
    attestation_variant = "azure-sev-snp"
    maa_url = "https://www.example.com" #  need to set this value
    image_version = "vX.Y.Z"  # defaults to provider version when not set
}

data "constellation_image" "img" {
    csp = "azure"
    attestation_variant = "azure-sev-snp"
    image_version = "vX.Y.Z"  # defaults to provider version when not set
}

module "azure_iam" {
  source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/iam/azure" // replace <version> with a Constellation version, e.g., v2.13.0
  region                 = var.location
  service_principal_name = var.service_principal_name
  resource_group_name    = var.resource_group_name
}

module "azure" {
  source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/azure" // replace <version> with a Constellation version, e.g., v2.13.0
  name                   = var.name
  user_assigned_identity = module.azure_iam.uami_id
  node_groups            = var.node_groups
  location               = var.location
  image_id               = module.fetch_image.image
  debug                  = var.debug
  resource_group         = module.azure_iam.base_resource_group
  create_maa             = var.create_maa
}
resource "constellation_cluster" "foo" {
   csp                                = "azure"
  constellation_microservice_version = "vX.Y.Z"
  name                               = module.azure.name
  uid                                = module.azure.uid
  image                              = data.constellation_image.img.reference
  attestation                        = data.constellation_attestation.att.attestation
  init_secret                        = module.azure.initSecret
  master_secret                      = ...
  master_secret_salt                 = ...
  measurement_salt                   = ...
  out_of_cluster_endpoint            = module.azure.out_of_cluster_endpoint
  in_cluster_endpoint                = module.azure.in_cluster_endpoint
  azure = {
    tenant_id                   = module.azure_iam.tenant_id
    subscription_id             = module.azure_iam.subscription_id
    uami                        = module.azure_iam.uami_id
    location                    = "northeurope"
    resource_group              = module.azure_iam.base_resource_group
    load_balancer_name          = module.azure.loadbalancer_name
    network_security_group_name = module.azure.network_security_group_name
  }
  network_config = {
    ip_cidr_node    = module.azure.ip_cidr_nodes
    ip_cidr_service = "10.96.0.0/12"
  }
}
}

  ```

  </tabItem>

  <tabItem value="aws" label="AWS">

  ```
  module "aws-constellation" {
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/aws-constellation" // replace <version> with a Constellation version, e.g., v2.13.0
    name        = "constell"
    zone        = "us-east-2c"
    name_prefix = "example"
    node_groups = {
      control_plane_default = {
        role          = "control-plane"
        zone          = "us-east-2c"
        instance_type = "m6a.xlarge"
        disk_size     = 30
        disk_type     = "gp3"
        initial_count = 3
      },
      worker_default = {
        role          = "worker"
        zone          = "us-east-2c"
        instance_type = "m6a.xlarge"
        disk_size     = 30
        disk_type     = "gp3"
        initial_count = 2
      }
    }
  }
  ```

  </tabItem>

  <tabItem value="gcp" label="GCP">

  ```
  module "gcp-constellation" {
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/gcp-constellation" // replace <version> with a Constellation version, e.g., v2.13.0
    name    = "constell"
    project = "constell-proj" // replace with your project id
    service_account_id = "constid"
    zone    = "europe-west2-a"
    node_groups = {
      control_plane_default = {
        role          = "control-plane"
        zone          = "europe-west2-a"
        instance_type = "n2d-standard-4"
        disk_size     = 30
        disk_type     = "pd-ssd"
        initial_count = 3
      },
      worker_default = {
        role          = "worker"
        zone          = "europe-west2-a"
        instance_type = "n2d-standard-4"
        disk_size     = 30
        disk_type     = "pd-ssd"
        initial_count = 2
      }
    }
  }
  ```

  </tabItem>
  </tabs>

1. Initialize and apply the file.
  ```bash
  terraform init
  terraform apply
  ```

## Custom setup
If you need custom infrastructure, you can download the Terraform module from the Constellation [GitHub releases](https://github.com/edgelesssys/constellation/releases) and modify them.
The module contains:
- `{csp}`: cluster infrastructure resources for the respective cloud provider
- `iam/{csp}`: contains the IAM resources used within the cluster

For cluster upgrades, please make sure to check the Constellation release notes for potential breaking changes in the infrastructure setup.

## Cluster upgrades
:::tip
For general information on cluster upgrades, see [Upgrade your cluster](./upgrade.md).
:::

First update the version of the Constellation Terraform provider. If you explicitly set versions (e.g. `image_version` or `constellation_microservice_version`), make sure to update them. Refer to [version support](https://github.com/edgelesssys/constellation/blob/main/dev-docs/workflows/versions-support.md), for more information on the version support policy.
Regarding the infrastructure / IAM modules, using a [remote address as module source](https://developer.hashicorp.com/terraform/language/modules/sources#fetching-archives-over-http) as shown in [Quick setup](#quick-setup) is recommended because it simplifies the upgrade process. For [local paths as module source](https://developer.hashicorp.com/terraform/language/modules/sources#local-paths), you would update the local files with the ones from the `terraform-module.zip` of the [Constellation release](https://github.com/edgelesssys/constellation/releases) and look out for potential breaking changes in the infrastructure setup.
The steps for applying the upgrade are as follows:

1. Update the `<version>` variable inside the `source` field of the module.
2. Upgrade the Terraform module and provider dependencies and apply the Constellation upgrade.
  ```bash
  terraform init -upgrade
  terraform apply
  ```
