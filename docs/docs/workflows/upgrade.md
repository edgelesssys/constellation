# Upgrade your cluster

Constellation provides an easy way to upgrade all components of your cluster, without disrupting it's availability.
Specifically, you can upgrade the Kubernetes version, the nodes' image, and the Constellation microservices.
You configure the desired versions in your local Constellation configuration and trigger upgrades with the `upgrade apply` command.
To learn about available versions you use the `upgrade check` command.
Which versions are available depends on the CLI version you are using.

:::caution

Upgrades arent't yet implemented for AWS. If you require this feature, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md)!

:::

## Update the CLI

Each CLI comes with a set of supported microservice and Kubernetes versions.
Most importantly, a given CLI version can only upgrade a cluster of the previous minor version, but not older ones.
This means that you have to upgrade your CLI and cluster one minor version at a time.

For example, if you are currently on CLI version v2.6 and the latest version is v2.8, you should
* upgrade the CLI to v2.7,
* upgrade the cluster to v2.7,
* and only then continue upgrading the CLI (and the cluster) to v2.8 after.

Also note that if your current Kubernetes version isn't supported by the next CLI version, use your current CLI to upgrade to a newer Kubernetes version first.

To learn which Kubernetes versions are supported by a particular CLI, run [constellation config kubernetes-versions](../reference/cli.md#constellation-config-kubernetes-versions).

## Migrate the configuration

The Constellation configuration file is located in the file `constellation-conf.yaml` in your workspace.
Refer to the [migration reference](../reference/config-migration.md) to check if you need to update fields in your configuration file.

## Check for upgrades

To learn which versions the current CLI can upgrade to and what's installed in your cluster, run:

```bash
constellation upgrade check
```

You can either enter the reported target versions into your config manually or run the above command with the `--write-config` flag.
When using this flag, the `kubernetesVersion`, `image`, and `microserviceVersion` fields are overwritten with the smallest available upgrade.

## Apply the upgrade

Once you updated your config with the desired versions, you can trigger the upgrade with this command:

```bash
constellation upgrade apply
```

Microservice upgrades will be finished within a few minutes, depending on the cluster size.
If you are interested, you can monitor pods restarting in the `kube-system` namespace with your tool of choice.

Image and Kubernetes upgrades take longer.
For each node in your cluster, a new node has to be created and joined.
The process usually takes up to ten minutes per node.
