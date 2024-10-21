# Use the Terraform provider

The Constellation Terraform provider allows to manage the full lifecycle of a Constellation cluster (namely creation, upgrades, and deletion) via Terraform.
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

2. Use one of the [example configurations for using the Constellation Terraform provider](https://github.com/edgelesssys/constellation/tree/main/terraform-provider-constellation/examples/full) or create a `main.tf` file and fill it with the resources you want to create. The [Constellation Terraform provider documentation](https://registry.terraform.io/providers/edgelesssys/constellation/latest) offers thorough documentation on the resources and their attributes.
3. Initialize and apply the Terraform configuration.
  <Tabs groupId="csp">

  <TabItem value="aws" label="AWS">
  Initialize the providers and apply the configuration.

  ```bash
  terraform init
  terraform apply
  ```

  Optionally, you can prefix the `terraform apply` command with `TF_LOG=INFO` to collect [Terraform logs](https://developer.hashicorp.com/terraform/internals/debugging) while applying the configuration. This may provide helpful output in debugging scenarios.
  </TabItem>
  <TabItem value="azure" label="Azure">
  When creating a cluster on Azure, you need to manually patch the policy of the MAA provider before creating the Constellation cluster, as this feature isn't available in Azure's Terraform provider yet. The Constellation CLI provides a utility for patching, but you
  can also do it manually.

  ```bash
  terraform init
  terraform apply -target module.azure_iam # adjust resource path if not using the example configuration
  terraform apply -target module.azure_infrastructure # adjust resource path if not using the example configuration
  constellation maa-patch $(terraform output -raw maa_url) # adjust output path / input if not using the example configuration or manually patch the resource
  terraform apply -target constellation_cluster.azure_example # adjust resource path if not using the example configuration
  ```

  Optionally, you can prefix the `terraform apply` command with `TF_LOG=INFO` to collect [Terraform logs](https://developer.hashicorp.com/terraform/internals/debugging) while applying the configuration. This may provide helpful output in debugging scenarios.

  Use the following policy if manually performing the patch.

  ```
  version= 1.0;
  authorizationrules
  {
      [type=="x-ms-azurevm-default-securebootkeysvalidated", value==false] => deny();
      [type=="x-ms-azurevm-debuggersdisabled", value==false] => deny();
      // The line below was edited to use the MAA provider within Constellation. Do not edit manually.
      //[type=="secureboot", value==false] => deny();
      [type=="x-ms-azurevm-signingdisabled", value==false] => deny();
      [type=="x-ms-azurevm-dbvalidated", value==false] => deny();
      [type=="x-ms-azurevm-dbxvalidated", value==false] => deny();
      => permit();
  };
  issuancerules
  {
  };
  ```

  </TabItem>
  <TabItem value="gcp" label="GCP">
  Initialize the providers and apply the configuration.

  ```bash
  terraform init
  terraform apply
  ```

  Optionally, you can prefix the `terraform apply` command with `TF_LOG=INFO` to collect [Terraform logs](https://developer.hashicorp.com/terraform/internals/debugging) while applying the configuration. This may provide helpful output in debugging scenarios.
  </TabItem>
  <TabItem value="stackit" label="STACKIT">
  Initialize the providers and apply the configuration.

  ```bash
  terraform init
  terraform apply
  ```

  Optionally, you can prefix the `terraform apply` command with `TF_LOG=INFO` to collect [Terraform logs](https://developer.hashicorp.com/terraform/internals/debugging) while applying the configuration. This may provide helpful output in debugging scenarios.
  </TabItem>

  </Tabs>
4. Connect to the cluster.

  ```bash
  terraform output -raw kubeconfig > constellation-admin.conf
  export KUBECONFIG=$(realpath constellation-admin.conf)
  ```

## Bringing your own infrastructure

Instead of using the example infrastructure used in the [quick setup](#quick-setup), you can also provide your own infrastructure.
If you need a starting point for a custom infrastructure setup, you can download the infrastructure / IAM Terraform modules for the respective CSP from the Constellation [GitHub releases](https://github.com/edgelesssys/constellation/releases). You can modify and extend the modules per your requirements, while keeping the basic functionality intact.
The module contains:

- `{csp}`: cloud resources the cluster runs on
- `iam/{csp}`: IAM resources used within the cluster

When upgrading your cluster, make sure to check the Constellation release notes for potential breaking changes in the reference infrastructure / IAM modules that need to be considered.

## Cluster upgrades

:::tip
Also see the [general documentation on cluster upgrades](./upgrade.md).
:::

The steps for applying the upgrade are as follows:

1. Update the version constraint of the Constellation Terraform provider in the `required_providers` block in your Terraform configuration.
2. If you explicitly set any of the version attributes of the provider's resources and data sources (e.g. `image_version` or `constellation_microservice_version`), make sure to update them too. Refer to Constellation's [version support policy](https://github.com/edgelesssys/constellation/blob/main/dev-docs/workflows/versions-support.md) for more information on how each Constellation version and its dependencies are supported.
3. Update the IAM / infrastructure configuration.
   - For [remote addresses as module sources](https://developer.hashicorp.com/terraform/language/modules/sources#fetching-archives-over-http), update the version number inside the address of the `source` field of the infrastructure / IAM module to the target version.
   - For [local paths as module sources](https://developer.hashicorp.com/terraform/language/modules/sources#local-paths) or when [providing your own infrastructure](#bringing-your-own-infrastructure), see the changes made in the reference modules since the upgrade's origin version and adjust your infrastructure / IAM configuration accordingly.
4. Upgrade the Terraform module and provider dependencies and apply the targeted configuration.

```bash
  terraform init -upgrade
  terraform apply
```
