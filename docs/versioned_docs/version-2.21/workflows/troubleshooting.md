# Troubleshooting

This section aids you in finding problems when working with Constellation.

## Common issues

### Issues with creating new clusters

When you create a new cluster, you should always use the [latest release](https://github.com/edgelesssys/constellation/releases/latest).
If something doesn't work, check out the [known issues](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22).

### Azure: Resource Providers can't be registered

On Azure, you may receive the following error when running `apply` or `terminate` with limited IAM permissions:

```shell-session
Error: Error ensuring Resource Providers are registered.

Terraform automatically attempts to register the Resource Providers it supports to
ensure it's able to provision resources.

If you don't have permission to register Resource Providers you may wish to use the
"skip_provider_registration" flag in the Provider block to disable this functionality.

[...]
```

To continue, please ensure that the [required resource providers](../getting-started/install.md#required-permissions) have been registered in your subscription by your administrator.

Afterward, set `ARM_SKIP_PROVIDER_REGISTRATION=true` as an environment variable and either run `apply` or `terminate` again.
For example:

```bash
ARM_SKIP_PROVIDER_REGISTRATION=true constellation apply
```

Or alternatively, for `terminate`:

```bash
ARM_SKIP_PROVIDER_REGISTRATION=true constellation terminate
```

### Azure: Can't update attestation policy

On Azure, you may receive the following error when running `apply` from within an Azure environment, e.g., an Azure VM:

```shell-session
An error occurred: patching policies: updating attestation policy: unexpected status code: 403 Forbidden
```

The problem occurs because the Azure SDK we use internally attempts to [authenticate towards the Azure API with the managed identity of your current environment instead of the Azure CLI token](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#DefaultAzureCredential).

We decided not to deviate from this behavior and comply with the ordering of credentials.

A solution is to add the [required permissions](../getting-started/install.md#required-permissions) to the managed identity of your environment. For example, the managed identity of your Azure VM, instead of the account that you've authenticated with in the Azure CLI.

If your setup requires a change in the ordering of credentials, please open an issue and explain your desired behavior.



### Nodes fail to join with error `untrusted measurement value`

This error indicates that a node's [attestation statement](../architecture/attestation.md) contains measurements that don't match the trusted values expected by the [JoinService](../architecture/microservices.md#joinservice).
This may for example happen if the cloud provider updates the VM's firmware such that it influences the [runtime measurements](../architecture/attestation.md#runtime-measurements) in an unforeseen way.
A failed upgrade due to an erroneous attestation config can also cause this error.
You can change the expected measurements to resolve the failure.

:::caution

Attestation and trusted measurements are crucial for the security of your cluster.
Be extra careful when manually changing these settings.
When in doubt, check if the encountered [issue is known](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22) or [contact support](https://github.com/edgelesssys/constellation#support).

:::

:::tip

During an upgrade with modified attestation config, a backup of the current configuration is stored in the `join-config` config map in the `kube-system` namespace under the `attestationConfig_backup` key. To restore the old attestation config after a failed upgrade, replace the value of `attestationConfig` with the value from `attestationConfig_backup`:

```bash
kubectl patch configmaps -n kube-system join-config -p "{\"data\":{\"attestationConfig\":\"$(kubectl get configmaps -n kube-system join-config -o "jsonpath={.data.attestationConfig_backup}")\"}}"
```

:::

You can use the `apply` command to change measurements of a running cluster:

1. Modify the `measurements` key in your local `constellation-conf.yaml` to the expected values.
2. Run `constellation apply`.

Keep in mind that running `apply` also applies any version changes from your config to the cluster.

You can run these commands to learn about the versions currently configured in the cluster:

- Kubernetes API server version: `kubectl get nodeversion constellation-version -o json -n kube-system | jq .spec.kubernetesClusterVersion`
- image version: `kubectl get nodeversion constellation-version -o json -n kube-system | jq .spec.imageVersion`
- microservices versions: `helm list --filter 'constellation-services' -n kube-system`

### Upgrading Kubernetes resources fails

Constellation manages its Kubernetes resources using Helm.
When applying an upgrade, the charts that are about to be installed, and a values override file `overrides.yaml`,
are saved to disk in your current workspace under `constellation-upgrade/upgrade-<timestamp>/helm-charts/`.
If upgrading the charts using the Constellation CLI fails, you can review these charts and try to manually apply the upgrade.

:::caution

Changing and manually applying the charts may destroy cluster resources and can lead to broken Constellation deployments.
Proceed with caution and when in doubt,
check if the encountered [issue is known](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22) or [contact support](https://github.com/edgelesssys/constellation#support).

:::

## Diagnosing issues

### Logs

To get started on diagnosing issues with Constellation, it's often helpful to collect logs from nodes, pods, or other resources in the cluster. Most logs are available through Kubernetes' standard
[logging interfaces](https://kubernetes.io/docs/concepts/cluster-administration/logging/).

To debug issues occurring at boot time of the nodes, you can use the serial console interface of the CSP while the machine boots to get a read-only view of the boot logs.

Apart from that, Constellation also offers further [observability integrations](../architecture/observability.md).

### Node shell access

Debugging via a shell on a node is [directly supported by Kubernetes](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#node-shell-session).

1. Figure out which node to connect to:

    ```bash
    kubectl get nodes
    # or to see more information, such as IPs:
    kubectl get nodes -o wide
    ```

2. Connect to the node:

    ```bash
    kubectl debug node/constell-worker-xksa0-000000 -it --image=busybox
    ```

    You will be presented with a prompt.

    The nodes file system is mounted at `/host`.

3. Once finished, clean up the debug pod:

    ```bash
    kubectl delete pod node-debugger-constell-worker-xksa0-000000-bjthj
    ```
