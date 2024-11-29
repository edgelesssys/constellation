# Troubleshooting

This section aids you in finding problems when working with Constellation.

## Azure: Resource Providers can't be registered

On Azure, you may receive the following error when running `create` or `terminate` with limited IAM permissions:

```shell-session
Error: Error ensuring Resource Providers are registered.

Terraform automatically attempts to register the Resource Providers it supports to
ensure it's able to provision resources.

If you don't have permission to register Resource Providers you may wish to use the
"skip_provider_registration" flag in the Provider block to disable this functionality.

[...]
```

To continue, please ensure that the [required resource providers](../getting-started/install.md#required-permissions) have been registered in your subscription by your administrator.

Afterward, set `ARM_SKIP_PROVIDER_REGISTRATION=true` as an environment variable and either run `create` or `terminate` again.
For example:

```bash
ARM_SKIP_PROVIDER_REGISTRATION=true constellation create --control-plane-nodes 1 --worker-nodes 2 -y
```

Or alternatively, for `terminate`:

```bash
ARM_SKIP_PROVIDER_REGISTRATION=true constellation terminate
```

## Cloud logging

To provide information during early stages of the node's boot process, Constellation logs messages into the cloud providers' log systems. Since these offerings **aren't** confidential, only generic information without any sensitive values are stored. This provides administrators with a high level understanding of the current state of a node.

You can view these information in the follow places:

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

1. In your Azure subscription find the Constellation resource group.
2. Inside the resource group find the Application Insights resource called `constellation-insights-*`.
3. On the left-hand side go to `Logs`, which is located in the section `Monitoring`.
    + Close the Queries page if it pops up.
5. In the query text field type in `traces`, and click `Run`.

To **find the disk UUIDs** use the following query: `traces | where message contains "Disk UUID"`

</TabItem>
<TabItem value="gcp" label="GCP">

1. Select the project that hosts Constellation.
2. Go to the `Compute Engine` service.
3. On the right-hand side of a VM entry select `More Actions` (a stacked ellipsis)
    + Select `View logs`

To **find the disk UUIDs** use the following query: `resource.type="gce_instance" text_payload=~"Disk UUID:.*\n" logName=~".*/constellation-boot-log"`

:::info

Constellation uses the default bucket to store logs. Its [default retention period is 30 days](https://cloud.google.com/logging/quotas#logs_retention_periods).

:::

</TabItem>
<TabItem value="aws" label="AWS">

1. Open [AWS CloudWatch](https://console.aws.amazon.com/cloudwatch/home)
2. Select [Log Groups](https://console.aws.amazon.com/cloudwatch/home#logsV2:log-groups)
3. Select the log group that matches the name of your cluster.
4. Select the log stream for control or worker type nodes.

</TabItem>
</Tabs>

## Connect to nodes

Debugging via a shell on a node is [directly supported by Kubernetes](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#node-shell-session).

1. Figure out which node to connect to:

    ```sh
    kubectl get nodes
    # or to see more information, such as IPs:
    kubectl get nodes -o wide
    ```

2. Connect to the node:

    ```sh
    kubectl debug node/constell-worker-xksa0-000000 -it --image=busybox
    ```

    You will be presented with a prompt.

    The nodes file system is mounted at `/host`.

3. Once finished, clean up the debug pod:

    ```sh
    kubectl delete pod node-debugger-constell-worker-xksa0-000000-bjthj
    ```
