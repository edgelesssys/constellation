# Use the Terraform provider

The Constellation Terraform provider allows to manage the full lifecycle of a Constellation cluster (namely creation, upgrades, and deletion) via Terraform.
<!-- TODO(elchead): check link during release -->
The provider is available through the [Terraform registry](https://registry.terraform.io/providers/edgelesssys/constellation/latest) and is released in lock-step with Constellation releases.

## Prerequisites

- a Linux / Mac operating system (ARM64/AMD64)
- a Terraform installation of version `v1.4.4` or above

## Quick setup

This example shows how to set up a Constellation cluster with the reference IAM and infrastructure setup. This setup is also used when creating a Constellation cluster through the Constellation CLI. You can either consume the IAM / infrastructure modules through a remote source (recommended) or local files. The latter requires downloading the infrastructure and IAM modules for the corresponding CSP from `terraform-modules.zip` on the [Constellation release page](https://github.com/edgelesssys/constellation/releases/latest) and placing them in the Terraform workspace directory.

1. Create a directory (workspace) for your Constellation cluster.

  ```bash
  mkdir constellation-workspace
  cd constellation-workspace
  ```

1. Create a `main.tf` file.
<!--TODO(elchead): AB#3607 put correct examples, with follow up PR with #2713 examples
  <tabs groupId="csp">

  <tabItem value="azure" label="Azure">
  </tabItem>

  <tabItem value="aws" label="AWS">
  </tabItem>

  <tabItem value="gcp" label="GCP">
  </tabItem>
  </tabs>-->

1. Initialize and apply the file.

  ```bash
  terraform init
  terraform apply
  ```

## Bringing your own infrastructure

If you need a custom infrastructure setup, you can download the infrastructure / IAM Terraform modules for the respective CSP from the Constellation [GitHub releases](https://github.com/edgelesssys/constellation/releases). You can modify / extend the modules, per your requirements, while keeping the basic functionality intact.
The module contains:

- `{csp}`: cloud resources the cluster runs on
- `iam/{csp}`: IAM resources used within the cluster

When upgrading your cluster, make sure to check the Constellation release notes for potential breaking changes in the reference infrastructure / IAM modules that need to be considered.

## Cluster upgrades

:::tip
For general information on cluster upgrades, see the [dedicated upgrade page](./upgrade.md).
:::

The steps for applying the upgrade are as follows:

1. Update the version constraint of the Constellation Terraform provider in the `required_providers` block in your Terraform configuration.
2. If you explicitly set any of the version attributes of the provider's resources and data sources (e.g. `image_version` or `constellation_microservice_version`), make sure to update them too. Refer to the [version support policy](https://github.com/edgelesssys/constellation/blob/main/dev-docs/workflows/versions-support.md) for more information on how each Constellation version and its dependencies are supported.
3. Update the IAM / infrastructure modules.
   - For [remote address as module source](https://developer.hashicorp.com/terraform/language/modules/sources#fetching-archives-over-http), update the version number inside the address of the `source` field  of the infra / IAM module to the target version.
   - For [local paths as module source](https://developer.hashicorp.com/terraform/language/modules/sources#local-paths), see the changes made in the reference modules since the upgrade's origin version and adjust your infrastructure configuration accordingly.
4. Upgrade the Terraform module and provider dependencies and apply the targeted configuration.

```bash
  terraform init -upgrade
  terraform apply
```
