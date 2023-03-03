# Terminate your cluster

You can terminate your cluster using the CLI. For this, you need the Terraform state directory named `constellation-terraform` in the current directory.

:::danger

All ephemeral storage and state of your cluster will be lost. Make sure any data is safely stored in persistent storage. Constellation can recreate your cluster and the associated encryption keys, but won't  backup your application data automatically.

:::

<tabs groupId="provider">
<tabItem value="cli" label="CLI">
Terminate the cluster by running:

```bash
constellation terminate
```

Or without confirmation (e.g., for automation purposes):

```bash
constellation terminate --yes
```

This deletes all resources created by Constellation in your cloud environment.
All local files created by the `create` and `init` commands are deleted as well, except for `constellation-mastersecret.json` and the configuration file.

:::caution

Termination can fail if additional resources have been created that depend on the ones managed by Constellation. In this case, you need to delete these additional
resources manually. Just run the `terminate` command again afterward to continue the termination process of the cluster.

:::

</tabItem>
<tabItem value="terraform" label="Terraform">
Terminate the cluster by running:

```bash
terraform destroy
```

Delete all files that are no longer needed:

```bash
rm constellation-id.json constellation-admin.conf
```

Only the `constellation-mastersecret.json` and the configuration file remain.

</tabItem>
</tabs>
