# Terraform Usage

[Terraform](https://www.terraform.io/) is an open-source Infrastructure as Code (IaC) framework which is being used by multiple Constellation components to manage cloud resources. This page describes our policy on using Terraform in Constellation.

:::info
This page assumes familiarity with Terraform. Refer to the [Terraform documentation](https://developer.hashicorp.com/terraform/docs) for an introduction.
:::

## Interacting with Terraform manually

Manual interaction with Terraform state created by Constellation (i.e. via the Terraform CLI) should only be performed by experienced users and only if absolutely necessary, as it may lead to unrecoverable loss of cloud resources. For the vast majority of users and use-cases, the interaction done by the [Constellation CLI](cli.md) is sufficient.

## Terraform state files

Constellation keeps Terraform state files in subdirectories together with the corresponding Terraform configuration files and metadata. When first performing an action in the Constellation CLI which uses Terraform internally, the needed subdirectory will be created.

Currently, these subdirectories are:

* `constellation-terraform` - Terraform state files for the resources used for the Constellation cluster
* `constellation-iam-terraform` - Terraform state files for IAM configuration

When working with either of the files, i.e. when running `constellation terminate` to [terminate a cluster](../workflows/terminate.md) (and delete it's cloud resources), the `constellation terraform` subdirectory needs to be in the current working directory. The same applies to the `constellation-iam-terraform` subdirectory when working with IAM configuration. The state directories shouldn't be deleted manually.
