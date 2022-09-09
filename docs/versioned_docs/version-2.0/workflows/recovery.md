# Recovery

Recovery of a Constellation cluster means getting a cluster back into a healthy state after it became unhealthy due to the underlying infrastructure.
Reasons for an unhealthy cluster can vary from a power outage, or planned reboot, to migration of nodes and regions.
Constellation keeps all stateful data protected and encrypted in a [stateful disk](../architecture/images.md#stateful-disk) attached to each node.
The stateful disk will be persisted across reboots.
The data restored from that disk contains the entire Kubernetes state including the application deployments.
Meaning after a successful recovery procedure the applications can continue operating without redeploying everything from scratch.

Recovery events are rare because Constellation is built for high availability and contains mechanisms to automatically replace and join nodes to the cluster.
Once a node reboots, the [*Bootstrapper*](../architecture/components.md#bootstrapper) will try to authenticate to the cluster's [*JoinService*](../architecture/components.md#joinservice) using remote attestation.
If successful the *JoinService* will return the encryption key for the stateful disk as part of the initialization response.
This process ensures that Constellation nodes can securely recover and rejoin a cluster autonomously.

In case of a disaster, where the control plane itself becomes unhealthy, Constellation provides a mechanism to recover that cluster and bring it back into a healthy state.
The `constellation recover` command connects to a node, establishes a secure connection using [attested TLS](../architecture/attestation.md#attested-tls-atls), and provides that node with the key to decrypt its stateful disk and continue booting.
This process has to be repeated until enough nodes are back running for establishing a [member quorum for etcd](https://etcd.io/docs/v3.5/faq/#what-is-failure-tolerance) and the Kubernetes state can be recovered.

## Identify unhealthy clusters

The first step to recovery is identifying when a cluster becomes unhealthy.
Usually, that's first observed when the Kubernetes API server becomes unresponsive.
The causes can vary but are often related to issues in the underlying infrastructure.
Recovery in Constellation becomes necessary if not enough control-plane nodes are in a healthy state to keep the control plane operational.

The health status of the Constellation nodes can be checked and monitored via the cloud service provider.
Constellation provides logging information on the boot process and status via [cloud logging](troubleshooting.md#cloud-logging).
In the following, you'll find detailed descriptions for identifying clusters stuck in recovery for each cloud environment.
Once you've identified that your cluster is in an unhealthy state you can use the [recovery](recovery.md#recover-your-cluster) command of the Constellation CLI to restore it.

<tabs>
<tabItem value="azure" label="Azure" default>

In the Azure cloud portal find the cluster's resource group `<cluster-name>-<suffix>`
Inside the resource group check that the control plane *Virtual machine scale set* `constellation-scale-set-controlplanes-<suffix>` has enough members in a *Running* state.
Open the scale set details page, on the left go to `Settings -> Instances` and check the *Status* field.

Second, check the boot logs of these *Instances*.
In the scale set's *Instances* view, open the details page of the desired instance.
Check the serial console output of that instance.
On the left open the *"Support + troubleshooting" -> "Serial console"* page:

In the serial console output search for `Waiting for decryption key`.
Similar output to the following means your node was restarted and needs to decrypt the [state disk](../architecture/images.md#state-disk):

```shell
{"level":"INFO","ts":"2022-08-01T08:02:20Z","caller":"cmd/main.go:46","msg":"Starting disk-mapper","version":"0.0.0","cloudProvider":"azure"}
{"level":"INFO","ts":"2022-08-01T08:02:20Z","logger":"setupManager","caller":"setup/setup.go:57","msg":"Preparing existing state disk"}
{"level":"INFO","ts":"2022-08-01T08:02:20Z","logger":"keyService","caller":"keyservice/keyservice.go:92","msg":"Waiting for decryption key. Listening on: [::]:9000"}
```

The node will then try to connect to the [*JoinService*](../architecture/components.md#joinservice) and obtain the decryption key.
If that fails, because the control plane is unhealthy, you will see log messages similar to the following:

```shell
{"level":"INFO","ts":"2022-08-01T08:02:21Z","logger":"keyService","caller":"keyservice/keyservice.go:118","msg":"Received list with JoinService endpoints: [10.9.0.5:30090 10.9.0.6:30090 10.9.0.7:30090 10.9.0.8:30090 10.9.0.9:30090 10.9.0.10:30090 10.9.0.11:30090 10.9.0.12:30090 10.9.0.13:30090 10.9.0.14:30090 10.9.0.15:30090 10.9.0.16:30090 10.9.0.17:30090 10.9.0.18:30090 10.9.0.19:30090 10.9.0.20:30090 10.9.0.21:30090 10.9.0.22:30090 10.9.0.23:30090]"}
{"level":"INFO","ts":"2022-08-01T08:02:21Z","logger":"keyService","caller":"keyservice/keyservice.go:145","msg":"Requesting rejoin ticket","endpoint":"10.9.0.5:30090"}
{"level":"ERROR","ts":"2022-08-01T08:02:21Z","logger":"keyService","caller":"keyservice/keyservice.go:148","msg":"Failed to request rejoin ticket","error":"rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 10.9.0.5:30090: connect: connection refused\"","endpoint":"10.9.0.5:30090"}
```

That means you have to recover that node manually.
Before you continue with the [recovery process](#recover-your-cluster) you need to know the node's IP address and state disk's UUID.
For the IP address, return to the instances *Overview* page and find the *Private IP address*.
For the UUID open the [Cloud logging](troubleshooting.md#azure) explorer.
Type `traces | where message contains "Disk UUID"` and click `Run`.
Find the entry corresponding to that instance `{"instance-name":"<cluster-name>-control-plane-<suffix>"}` and take the UUID from the message field `Disk UUID: <UUID>`.

</tabItem>
<tabItem value="gcp" label="GCP" default>

First, check that the control plane *Instance Group* has enough members in a *Ready* state.
Go to *Instance Groups* and check the group for the cluster's control plane `<cluster-name>-control-plane-<suffix>`.

Second, check the status of the *VM Instances*.
Go to *VM Instances* and open the details of the desired instance.
Check the serial console output of that instance by opening the *logs -> "Serial port 1 (console)"* page:

![GCP portal serial console link](../_media/recovery-gcp-serial-console-link.png)

In the serial console output search for `Waiting for decryption key`.
Similar output to the following means your node was restarted and needs to decrypt the [state disk](../architecture/images.md#state-disk):

```shell
{"level":"INFO","ts":"2022-07-29T09:45:55Z","caller":"cmd/main.go:46","msg":"Starting disk-mapper","version":"0.0.0","cloudProvider":"gcp"}
{"level":"INFO","ts":"2022-07-29T09:45:55Z","logger":"setupManager","caller":"setup/setup.go:57","msg":"Preparing existing state disk"}
{"level":"INFO","ts":"2022-07-29T09:45:55Z","logger":"keyService","caller":"keyservice/keyservice.go:92","msg":"Waiting for decryption key. Listening on: [::]:9000"}
```

The node will then try to connect to the [*JoinService*](../architecture/components.md#joinservice) and obtain the decryption key.
If that fails, because the control plane is unhealthy, you will see log messages similar to the following:

```shell
{"level":"INFO","ts":"2022-07-29T09:46:15Z","logger":"keyService","caller":"keyservice/keyservice.go:118","msg":"Received list with JoinService endpoints: [192.168.178.2:30090]"}
{"level":"INFO","ts":"2022-07-29T09:46:15Z","logger":"keyService","caller":"keyservice/keyservice.go:145","msg":"Requesting rejoin ticket","endpoint":"192.168.178.2:30090"}
{"level":"ERROR","ts":"2022-07-29T09:46:15Z","logger":"keyService","caller":"keyservice/keyservice.go:148","msg":"Failed to request rejoin ticket","error":"rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 192.168.178.2:30090: connect: connection refused\"","endpoint":"192.168.178.2:30090"}
```

That means you have to recover that node manually.
Before you continue with the [recovery process](#recover-your-cluster) you need to know the node's IP address and state disk's UUID.
For the IP address go to the *"VM Instance" -> "network interfaces"* page and take the address from *"Primary internal IP address."*
For the UUID open the [Cloud logging](troubleshooting.md#cloud-logging) explorer, you'll find that right above the serial console link (see the picture above).
Search for `Disk UUID: <UUID>`.

</tabItem>
</tabs>

## Recover your cluster

Depending on the size of your cluster and the number of unhealthy control plane nodes the following process needs to be repeated until a [member quorum for etcd](https://etcd.io/docs/v3.5/faq/#what-is-failure-tolerance) is established.
For example, assume you have 5 control-plane nodes in your cluster and 4 of them have been rebooted due to a maintenance downtime in the cloud environment.
You have to run through the following process for 2 of these nodes and recover them manually to recover the quorum.
From there, your cluster will auto heal the remaining 2 control-plane nodes and the rest of your cluster.

Recovering a node requires the following parameters:

* The node's IP address
* The node's state disk UUID
* Access to the master secret of the cluster

See the [Identify unhealthy clusters](#identify-unhealthy-clusters) description of how to obtain the node's IP address and state disk UUID.
Note that the recovery command needs to connect to the recovering nodes.
Nodes only have private IP addresses in the VPC of the cluster, hence, the command needs to be issued from within the VPC network of the cluster.
The easiest approach is to set up a jump host connected to the VPC network and perform the recovery from there.

Given these prerequisites a node can be recovered like this:

```bash
$ constellation recover -e 34.107.89.208 --disk-uuid b27f817c-6799-4c0d-81d8-57abc8386b70 --master-secret constellation-mastersecret.json
Pushed recovery key.
```

In the serial console output of the node you'll see a similar output to the following:

```shell
[ 3225.621753] EXT4-fs (dm-1): INFO: recovery required on readonly filesystem
[ 3225.628807] EXT4-fs (dm-1): write access will be enabled during recovery
[ 3226.295816] EXT4-fs (dm-1): recovery complete
[ 3226.301618] EXT4-fs (dm-1): mounted filesystem with ordered data mode. Opts: (null). Quota mode: none.
[ 3226.338157] systemd[1]: run-state.mount: Deactivated successfully.
[[0;32m  OK  [[ 3226.347833] systemd[1]: Finished Prepare encrypted state disk.
0m] Finished [0;1;39mPrepare encrypted state disk[0m.
         Startin[ 3226.363705] systemd[1]: Starting OSTree Prepare OS/...
g [0;1;39mOSTre[ 3226.370625] ostree-prepare-root[939]: preparing sysroot at /sysroot
e Prepare OS/[0m...
```

After enough control plane nodes have been recovered and the Kubernetes cluster becomes healthy again, the rest of the cluster will start auto healing using the mechanism described above.
