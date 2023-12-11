# Use the Terraform provider
You can manage a Constellation cluster through Terraform.
<!-- TODO(elchead): check link during release -->
The provider is available through the [Terraform registry](https://registry.terraform.io/providers/edgelesssys/constellation/latest) and is released as part of Constellation releases. This page shows how to use the Constellation Terraform provider.

## Prerequisites
- a Linux / Mac operating system (ARM64/AMD64)
- a Terraform installation of version `v1.4.4` or above

## Quick setup
The example shows how to set up a Constellation cluster with the default infrastructure and IAM setup. It's easiest to consume the module through a remote source, as shown below.
This allows to upgrade the cluster to a newer Constellation version by simply updating the module source. When using custom infrastructure, make sure to check the Constellation release notes for potential breaking changes in the infrastructure setup.


1. Create a directory (workspace) for your Constellation cluster.
  ```bash
  mkdir constellation-workspace
  cd constellation-workspace
  ```

1. Create a `main.tf` file to call the CSP specific Constellation module.

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
    attestation_variant = "azure-sev-snp"
    maa_url = "https://www.example.com" #  need to set this value!
}

data "constellation_image" "img" {
    attestation_variant = "azure-sev-snp"
    csp = "azure"
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
    uid = "bar"
    name = "baz"
    image = data.constellation_image.ref # or provide manually crafted values
    kubernetes_version = "v1.27.6"
    init_endpoint = "10.10.10.10" # should use public ip of LB resource, ideally also provisioned through TF
    kubernetes_api_endpoint = "10.10.10.10" # should use public ip of LB resource, ideally also provisioned through TF
    constellation_microservice_version = "v2.13.0" # optional value, set to provider version by default.
    extra_microservices = {
        csi_driver = true
        # + more
        # possiblly also constellation microservices with version and maybe service options,
        # which would make constellation_microservice_version obsolete.
        # exact API TBD
    }
    master_secret = "foo" # updating this would force recreation of the cluster
    init_secret = "bar" # maybe derive from master_secret, updating this would force recreation of the cluster
    network_config = {
        # TBD
        # should contain CIDRs for pod network, service cidr, node network... for Cilium
        # the aforementioned values might be outputs of infrastructure that is also provisioned
        # through Terraform, such as a VPC.
        # and in-cluster Kubernetes API endpoint, e.g. for Kubelets
    }
    attestation = data.constellation_attestation.attestation # or provide manually crafted values
}

  <!--module "azure-constellation" {
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/azure-constellation" // replace <version> with a Constellation version, e.g., v2.13.0
    name = "constell"
    location = "northeurope"
    service_principal_name = "az-sp"
    resource_group_name = "constell-rg"
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
  }-->
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

Using a [remote address as module source](https://developer.hashicorp.com/terraform/language/modules/sources#fetching-archives-over-http) as shown in [Quick setup](#quick-setup) is recommended because it simplifies the upgrade process. For [local paths as module source](https://developer.hashicorp.com/terraform/language/modules/sources#local-paths), you would need to look out for potential breaking changes in the infrastructure setup.
The steps for the remote source setup are as follows:

1. Update the `<version>` variable inside the `source` field of the module.
2. Upgrade the Terraform module and provider dependencies and apply the Constellation upgrade.
  ```bash
  terraform init -upgrade
  terraform apply
  ```
