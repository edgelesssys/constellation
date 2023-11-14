# Use the Terraform module
You can manage a Constellation cluster through Terraform.
The module package is available as part of the [GitHub release](https://github.com/edgelesssys/constellation/releases/). It consists of a convenience module for each cloud service provider (`{csp}-constellation`) that combines the IAM (`infrastructure/{csp}/iam`), infrastructure (`infrastructure/{csp}`), and constellation (`constellation-cluster`) modules.

## Prerequisites
- a Linux / Mac operating system
- a Terraform installation of version `v1.4.4` or above

## Quick setup
The convenience module allows setting up a Constellation cluster with a single module. It's easiest to consume the module through a remote source, as shown below.
This allows to upgrade the cluster to a newer Constellation version by simply updating the module source.

1. Create a directory (workspace) for your Constellation cluster.
  ```bash
  mkdir constellation-workspace
  cd constellation-workspace
  ```

1. Create a `main.tf` file to call the CSP specific Constellation module.

  <tabs groupId="csp">

  <tabItem value="azure" label="Azure">

  ```
  module "azure-constellation" {
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/azure-constellation" // insert Constellation version here.
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
  }
  ```

  </tabItem>

  <tabItem value="aws" label="AWS">

  ```
  module "aws-constellation" {
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/aws-constellation" // insert Constellation version here.
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
    source = "https://github.com/edgelesssys/constellation/releases/download/<version>/terraform-module.zip//terraform-module/gcp-constellation" // insert Constellation version here.
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

3. Initialize and apply the module.
  ```bash
  terraform init
  terraform apply
  ```

## Custom setup
If you need to separate IAM and cluster management or need custom infrastructure, you can also call the submodules individually.
Look at the respective convenience module (`{csp}-constellation`) for how you can structure the module calls.
The submodules are:
- `constellation-cluster`: manages the Constellation cluster
- `fetch-image`: translates the Constellation image version to the image ID of the cloud service provider
- `infrastructure/{csp}`: contains the cluster infrastructure resources
- `infrastructure/iam/{csp}`: contains the IAM resources used within the cluster

## Cluster upgrades
:::tip
For general information on cluster upgrades, see [Upgrade your cluster](./upgrade.md).
:::

Using a [remote address as module source](https://developer.hashicorp.com/terraform/language/modules/sources#fetching-archives-over-http) as shown in [Quick setup](#quick-setup) is recommended because it simplifies the upgrade process. For [local paths as module source](https://developer.hashicorp.com/terraform/language/modules/sources#local-paths), you would need to manually overwrite the Terraform files in the Terraform workspace. The steps for the remote source setup are as follows:

1. Update the local `module_version` variable.
2. Upgrade the Terraform module and provider dependencies and apply the Constellation upgrade.
  ```bash
  terraform init -upgrade
  terraform apply
  ```
