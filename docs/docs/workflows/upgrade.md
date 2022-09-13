# Upgrade your cluster

Constellation provides an easy way to upgrade from one release to the next.
This involves choosing a new VM image to use for all nodes in the cluster and updating the cluster's expected measurements.

## Plan the upgrade

If you don't already know the image you want to upgrade to, use the `upgrade plan` command to pull in a list of available updates.

```bash
constellation upgrade plan
```

The command will let you interactively choose from a list of available updates and prepare your Constellation config file for the next step.

If you plan to use the command in scripts, use the `--file` flag to compile the available options into a YAML file.
You can then manually set the chosen upgrade option in your Constellation config file.

:::caution

The `constellation upgrade plan` will only work for official Edgeless release images.
If your cluster is using a custom image or a debug image, the Constellation CLI will fail to find compatible images.
However, you may still use the `upgrade execute` command by manually selecting a compatible image and setting it in your config file.

:::

## Execute the upgrade

Once your config file has been prepared with the new image and measurements, use the `upgrade execute` command to initiate the upgrade.

```bash
constellation upgrade execute
```

After the command has finished, the cluster will automatically replace old nodes using a rolling update strategy to ensure no downtime of the control or data plane.
