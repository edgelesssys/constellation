# Upgrade your cluster

Constellation provides an easy way to upgrade to the next release.
This involves updating the CLI, choosing a new VM image to use for all nodes in the cluster, and updating the cluster's expected measurements.

## Update the CLI

New features and bug fixes are added to the CLI with every release. To use them, update the CLI to the latest version by following the instructions in the [installation guide](../getting-started/install.md).

## Migrate the configuration

The Constellation configuration file is located in the file `constellation-conf.yaml` in your workspace.
Refer to the [migration reference](../reference/config-migration.md) to check if you need to update fields in your configuration file.

## Plan the upgrade

If you don't already know the image you want to upgrade to, use the `upgrade plan` command to pull a list of available updates.

```bash
constellation upgrade plan
```

The command lets you interactively choose from a list of available updates and prepares your Constellation config file for the next step.

To use the command in scripts, use the `--file` flag to compile the available options into a YAML file.
You can then set the chosen upgrade option in your Constellation config file.

:::caution

`constellation upgrade plan` only works for official Edgeless release images.
If your cluster is using a custom image, the Constellation CLI will fail to find compatible images.
However, you may still use the `upgrade execute` command by manually selecting a compatible image and setting it in your config file.

:::

## Execute the upgrade

Once your config file has been prepared with the new image and measurements, use the `upgrade execute` command to initiate the upgrade.

```bash
constellation upgrade execute
```

After the command has finished, the cluster will automatically replace old nodes using a rolling update strategy to ensure no downtime of the control or data plane.
