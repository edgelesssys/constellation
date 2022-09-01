# Terminate your cluster

You can terminate your cluster using the CLI.
You need the state file of your running cluster named `constellation-state.json` in the current directory.

:::danger

All ephemeral storage and state of your cluster will be lost. Make sure any data is safely stored in persistent storage. Constellation can recreate your cluster and the associated encryption keys, but won't  backup your application data automatically.

:::

Terminate the cluster by running:

```bash
constellation terminate
```

This deletes all resources created by Constellation in your cloud environment.
All local files created by the `create` and `init` commands are deleted as well, except the *master secret* `constellation-mastersecret.json` and the configuration file.

:::caution

Termination can fail if additional resources have been created that depend on the ones managed by Constellation. In this case, you need to delete these additional
resources manually. Just run the `terminate` command again afterward to continue the termination process of the cluster.

:::
