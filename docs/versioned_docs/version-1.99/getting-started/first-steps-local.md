# First steps with MiniConstellation

<!-- vale off -->
With the `constellation mini` command, you can deploy and test Constellation locally without a cloud subscription. This mode is called MiniConstellation. Conceptually, MiniConstellation is similar to [MicroK8s](https://microk8s.io/), [K3s](https://k3s.io/), and [minikube](https://minikube.sigs.k8s.io/docs/).
<!-- vale on -->

MiniConstellation uses virtualization to create a local cluster with one control-plane node and one worker node. It **doesn't** require hardware with Confidential VM (CVM) support. For attestation, MiniConstellation currently uses a software-based vTPM provided by KVM/QEMU.

:::caution

MiniConstellation has specific soft- and hardware requirements such as a Linux OS running on an x86-64 CPU. Pay attention to all [prerequisites](#prerequisites) when setting up.

:::

:::note

Since MiniConstellation runs on your local system, cloud features such as load balancing,
attaching persistent storage, or autoscaling aren't available.

:::

## Prerequisites

* A Linux OS with the following components installed
  * [Constellation CLI](./install.md#install-the-constellation-cli)
  * [KVM kernel module](https://www.linux-kvm.org/page/Main_Page)
  * [Docker](https://docs.docker.com/engine/install/)
  * [xsltproc](https://gitlab.gnome.org/GNOME/libxslt/-/wikis/home)
  * (Optional) [virsh](https://www.libvirt.org/manpages/virsh.html) to observe and access your nodes
* Other system requirements
  * An x86-64 CPU with at least 4 cores (6 cores are recommended)
  * At least 4 GB RAM (6 GB are recommended)
  * 20 GB of free disk space
  * Hardware virtualization enabled in the BIOS/UEFI (often referred to as Intel VT-x or AMD-V/SVM)
  * `iptables` rules configured to not drop forwarded packages.
    If running the following command returns no error, please follow [the troubleshooting guide](#vms-have-no-internet-access):

    ```bash
    sudo iptables -S | grep -q -- '-P FORWARD DROP'
    ```

## Create your cluster

The following creates your MiniConstellation cluster (may take up to 10 minutes to complete):

```bash
constellation mini up
```

This will configure your current directory as the [workspace](../architecture/orchestration.md#workspaces) for this cluster.
All `constellation` commands concerning this cluster need to be issued from this directory.

## Connect `kubectl`

Configure `kubectl` to connect to your local Constellation cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
```

Your cluster initially consists of a single control-plane node:

```shell-session
$ kubectl get nodes
NAME              STATUS   ROLES           AGE   VERSION
control-plane-0   Ready    control-plane   66s   v1.24.6
```

A worker node will request to join the cluster shortly. Before the new worker node is allowed to join the cluster, its state is verified using remote attestation by the [JoinService](../architecture/microservices.md#joinservice).
If verification passes successfully, the new node receives keys and certificates to join the cluster.

You can follow this process by viewing the logs of the JoinService:

```shell-session
$ kubectl logs -n kube-system daemonsets/join-service -f
{"level":"INFO","ts":"2022-10-14T09:32:20Z","caller":"cmd/main.go:48","msg":"Constellation Node Join Service","version":"2.1.0","cloudProvider":"qemu"}
{"level":"INFO","ts":"2022-10-14T09:32:20Z","logger":"validator","caller":"watcher/validator.go:96","msg":"Updating expected measurements"}
...
```

Once the worker node has joined your cluster, it may take a couple of minutes for all resources to become available.
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

Once you are done, you can clean up the created resources using the following command:

```bash
constellation mini down
```

This will destroy your cluster and clean up your workspace.
The VM image and cluster configuration file (`constellation-conf.yaml`) will be kept and may be reused to create new clusters.

## Troubleshooting

Make sure to use the [latest release](https://github.com/edgelesssys/constellation/releases/latest) and check out the [known issues](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22).

### VMs have no internet access

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
