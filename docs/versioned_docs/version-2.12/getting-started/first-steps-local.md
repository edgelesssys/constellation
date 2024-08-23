# First steps with a local cluster

A local cluster lets you deploy and test Constellation without a cloud subscription.
You have two options:

* Use MiniConstellation to automatically deploy a two-node cluster.
* For more fine-grained control, create the cluster using the QEMU provider.

Both options use virtualization to create a local cluster with control-plane nodes and worker nodes. They **don't** require hardware with Confidential VM (CVM) support. For attestation, they currently use a software-based vTPM provided by KVM/QEMU.

You need an x64 machine with a Linux OS.
You can use a VM, but it needs nested virtualization.

## Prerequisites

* Machine requirements:
  * An x86-64 CPU with at least 4 cores (6 cores are recommended)
  * At least 4 GB RAM (6 GB are recommended)
  * 20 GB of free disk space
  * Hardware virtualization enabled in the BIOS/UEFI (often referred to as Intel VT-x or AMD-V/SVM) / nested-virtualization support when using a VM
* Software requirements:
  * Linux OS with [KVM kernel module](https://www.linux-kvm.org/page/Main_Page)
    * Recommended: Ubuntu 22.04 LTS
  * [Docker](https://docs.docker.com/engine/install/)
  * [xsltproc](https://gitlab.gnome.org/GNOME/libxslt/-/wikis/home)
  * (Optional) [virsh](https://www.libvirt.org/manpages/virsh.html) to observe and access your nodes

### Software installation on Ubuntu

```bash
# install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install docker-ce
# install other dependencies
sudo apt install xsltproc
sudo snap install kubectl --classic
# install Constellation CLI
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64
sudo install constellation-linux-amd64 /usr/local/bin/constellation
# do not drop forwarded packages
sudo iptables -P FORWARD ACCEPT
```

## Create a cluster

<Tabs groupId="csp">
<TabItem value="mini" label="MiniConstellation">

<!-- vale off -->
With the `constellation mini` command, you can deploy and test Constellation locally. This mode is called MiniConstellation. Conceptually, MiniConstellation is similar to [MicroK8s](https://microk8s.io/), [K3s](https://k3s.io/), and [minikube](https://minikube.sigs.k8s.io/docs/).
<!-- vale on -->

:::caution

MiniConstellation has specific soft- and hardware requirements such as a Linux OS running on an x86-64 CPU. Pay attention to all [prerequisites](#prerequisites) when setting up.

:::

:::note

Since MiniConstellation runs on your local system, cloud features such as load balancing,
attaching persistent storage, or autoscaling aren't available.

:::

The following creates your MiniConstellation cluster (may take up to 10 minutes to complete):

```bash
constellation mini up
```

This will configure your current directory as the [workspace](../architecture/orchestration.md#workspaces) for this cluster.
All `constellation` commands concerning this cluster need to be issued from this directory.

</TabItem>
<TabItem value="qemu" label="QEMU">

With the QEMU provider, you can create a local Constellation cluster as if it were in the cloud. The provider uses [QEMU](https://www.qemu.org/) to create multiple VMs for the cluster nodes, which interact with each other.

:::caution

Constellation on QEMU has specific soft- and hardware requirements such as a Linux OS running on an x86-64 CPU. Pay attention to all [prerequisites](#prerequisites) when setting up.

:::

:::note

Since Constellation on QEMU runs on your local system, cloud features such as load balancing,
attaching persistent storage, or autoscaling aren't available.

:::

1. To set up your local cluster, you need to create a configuration file for Constellation first.

  ```bash
  constellation config generate qemu
  ```

  This creates a [configuration file](../workflows/config.md) for QEMU called `constellation-conf.yaml`. After that, your current folder also becomes your [workspace](../architecture/orchestration.md#workspaces). All `constellation` commands for your cluster need to be executed from this directory.

2. Now you can create your cluster and its nodes. `constellation create` uses the options set in `constellation-conf.yaml`.

  ```bash
  constellation create
  ```

  The Output should look like the following:

  ```shell-session
  $ constellation create
  Your Constellation cluster was created successfully.
  ```

3. Initialize the cluster

  ```bash
  constellation init
  ```

  This should give the following output:

  ```shell-session
  $ constellation init
  Your Constellation master secret was successfully written to ./constellation-mastersecret.json
  Note: If you just created the cluster, it can take a few minutes to connect.
  Initializing cluster ...
  Your Constellation cluster was successfully initialized.

  Constellation cluster identifier  g6iMP5wRU1b7mpOz2WEISlIYSfdAhB0oNaOg6XEwKFY=
  Kubernetes configuration          constellation-admin.conf

  You can now connect to your cluster by executing:
          export KUBECONFIG="$PWD/constellation-admin.conf"
  ```

  The cluster's identifier will be different in your output.
  Keep `constellation-mastersecret.json` somewhere safe.
  This will allow you to [recover your cluster](../workflows/recovery.md) in case of a disaster.

  :::info

  Depending on your setup, `constellation init` may take 10+ minutes to complete.

  :::

4. Configure kubectl

  ```bash
  export KUBECONFIG="$PWD/constellation-admin.conf"
  ```

</TabItem>
</Tabs>

## Connect to the cluster

Your cluster initially consists of a single control-plane node:

```shell-session
$ kubectl get nodes
NAME              STATUS   ROLES           AGE   VERSION
control-plane-0   Ready    control-plane   66s   v1.24.6
```

Additional nodes will request to join the cluster shortly. Before each additional node is allowed to join the cluster, its state is verified using remote attestation by the [JoinService](../architecture/microservices.md#joinservice).
If verification passes successfully, the new node receives keys and certificates to join the cluster.

You can follow this process by viewing the logs of the JoinService:

```shell-session
$ kubectl logs -n kube-system daemonsets/join-service -f
{"level":"INFO","ts":"2022-10-14T09:32:20Z","caller":"cmd/main.go:48","msg":"Constellation Node Join Service","version":"2.1.0","cloudProvider":"qemu"}
{"level":"INFO","ts":"2022-10-14T09:32:20Z","logger":"validator","caller":"watcher/validator.go:96","msg":"Updating expected measurements"}
...
```

Once all nodes have joined your cluster, it may take a couple of minutes for all resources to become available.
You can check on the state of your cluster by running the following:

```shell-session
$ kubectl get nodes
NAME              STATUS   ROLES           AGE     VERSION
control-plane-0   Ready    control-plane   2m59s   v1.24.6
worker-0          Ready    <none>          32s     v1.24.6
```

## Deploy a sample application

1. Deploy the [emojivoto app](https://github.com/BuoyantIO/emojivoto)

  ```bash
  kubectl apply -k github.com/BuoyantIO/emojivoto/kustomize/deployment
  ```

2. Expose the frontend service locally

  ```bash
  kubectl wait --for=condition=available --timeout=60s -n emojivoto --all deployments
  kubectl -n emojivoto port-forward svc/web-svc 8080:80 &
  curl http://localhost:8080
  kill %1
  ```

## Terminate your cluster

<Tabs groupId="csp">
<TabItem value="mini" label="MiniConstellation">

Once you are done, you can clean up the created resources using the following command:

```bash
constellation mini down
```

This will destroy your cluster and clean up your workspace.
The VM image and cluster configuration file (`constellation-conf.yaml`) will be kept and may be reused to create new clusters.

</TabItem>
<TabItem value="qemu" label="QEMU">

Once you are done, you can clean up the created resources using the following command:

```bash
constellation terminate
```

This should give the following output:

```shell-session
$ constellation terminate
You are about to terminate a Constellation cluster.
All of its associated resources will be DESTROYED.
This action is irreversible and ALL DATA WILL BE LOST.
Do you want to continue? [y/n]:
```

Confirm with `y` to terminate the cluster:

```shell-session
Terminating ...
Your Constellation cluster was terminated successfully.
```

This will destroy your cluster and clean up your workspace.
The VM image and cluster configuration file (`constellation-conf.yaml`) will be kept and may be reused to create new clusters.

</TabItem>
</Tabs>

## Troubleshooting

Make sure to use the [latest release](https://github.com/edgelesssys/constellation/releases/latest) and check out the [known issues](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22).

### VMs have no internet access / CLI remains in "Initializing cluster" state

`iptables` rules may prevent your VMs from accessing the internet.
Make sure your rules aren't dropping forwarded packages.

List your rules:

```bash
sudo iptables -S
```

The output may look similar to the following:

```shell-session
-P INPUT ACCEPT
-P FORWARD DROP
-P OUTPUT ACCEPT
-N DOCKER
-N DOCKER-ISOLATION-STAGE-1
-N DOCKER-ISOLATION-STAGE-2
-N DOCKER-USER
```

If your `FORWARD` chain is set to `DROP`, you need to update your rules:

```bash
sudo iptables -P FORWARD ACCEPT
```
