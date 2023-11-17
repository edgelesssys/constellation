# Terraform usage

[Terraform](https://www.terraform.io/) is an Infrastructure as Code (IaC) framework to manage cloud resources. This page explains how Constellation uses it internally and how advanced users may manually use it to have more control over the resource creation.

:::info
Information on this page is intended for users who are familiar with Terraform.
It's not required for common usage of Constellation.
See the [Terraform documentation](https://developer.hashicorp.com/terraform/docs) if you want to learn more about it.
:::

## Terraform state files

Constellation keeps Terraform state files in subdirectories of the workspace together with the corresponding Terraform configuration files and metadata.
The subdirectories are created on the first Constellation CLI action that uses Terraform internally.

Currently, these subdirectories are:

* `constellation-terraform` - Terraform state files for the resources of the Constellation cluster
* `constellation-iam-terraform` - Terraform state files for IAM configuration

As with all commands, commands that work with these files (e.g., `apply`, `terminate`, `iam`) have to be executed from the root of the cluster's [workspace directory](../architecture/orchestration.md#workspaces). You usually don't need and shouldn't manipulate or delete the subdirectories manually.

## Interacting with Terraform manually

Manual interaction with Terraform state created by Constellation (i.e., via the Terraform CLI) should only be performed by experienced users. It may lead to unrecoverable loss of cloud resources. For the majority of users and use cases, the interaction done by the [Constellation CLI](cli.md) is sufficient.

## Terraform debugging

To debug Terraform issues, the Constellation CLI offers the `tf-log` flag. You can set it to any of [Terraform's log levels](https://developer.hashicorp.com/terraform/internals/debugging):
* `JSON` (JSON-formatted logs at `TRACE` level)
* `TRACE`
* `DEBUG`
* `INFO`
* `WARN`
* `ERROR`

The log output is written to the `terraform.log` file in the workspace directory. The output is appended to the file on each run.
