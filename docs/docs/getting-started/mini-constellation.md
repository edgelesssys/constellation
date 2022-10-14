# MiniConstellation

With `constellation mini`, you can deploy and test Constellation locally without a cloud subscription.

The command uses virtualization to create a local cluster with one control-plane and one worker node.

:::info

Since MiniConstellation is running on your local system, common cloud features, such as load-balancing,
attaching persistent storage, or autoscaling, are unavailable.

:::

## Prerequisites

* [Constellation CLI](./install.md#install-the-constellation-cli)
* An x86-64 CPU with at least 4 cores
  * Recommended are 6 cores or more
* Hardware virtualization enabled in the BIOS/UEFI (often referred to as Intel VT-x or AMD-V/SVM)
* At least 4 GB RAM
  * Recommend are 6 GB or more
* 20 GB of free disk space
* A Linux operating system
* [KVM kernel module](https://www.linux-kvm.org/page/Main_Page) enabled
* [Docker](https://docs.docker.com/engine/install/)
* [xsltproc](https://gitlab.gnome.org/GNOME/libxslt/-/wikis/home)
  * Install on Ubuntu:

    ```bash
    sudo apt install xsltproc
    ```

  * Install on Fedora

    ```bash
    sudo dnf install xsltproc
    ```

* (Optional) [`virsh`](https://www.libvirt.org/manpages/virsh.html) to observe and access your nodes

## Create your cluster

Setting up your MiniConstellation cluster is as easy as running the following command:

```bash
constellation mini up
```

This will configure your current directory as the [workspace](../architecture/orchestration.md#workspaces) for this cluster.
All `constellation` commands concerning this cluster need to be issued from this directory.

The command will create your cluster and initialize it. Depending on your system, this may take up to 10 minutes.
The output should look like the following:

```shell-session
$ constellation mini up
Downloading image to ./constellation.qcow2
Done.

Creating cluster in QEMU ...
Cluster successfully created.
Connect to the VMs by executing:
        virsh -c qemu+tcp://localhost:16599/system

Your Constellation master secret was successfully written to ./constellation-mastersecret.json
Initializing cluster ...
Your Constellation cluster was successfully initialized.

Constellation cluster identifier  hmrRaTJEKHk6zlM6wcTCGxZ+7HAA16ec4T9CmKs12uQ=
Kubernetes configuration          constellation-admin.conf

You can now connect to your cluster by executing:
        export KUBECONFIG="$PWD/constellation-admin.conf"
```

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

A worker node will request to join the cluster shortly.

Before the new worker node is allowed to join the cluster, its state is verified using remote attestation by the [JoinService](../architecture/components.md#joinservice).
If verification passes successfully, the new node receives keys and certificates to join the cluster.

You can follow this process by viewing the logs of the JoinService:

```shell-session
$ kubectl logs -n kube-system daemonsets/join-service -f
{"level":"INFO","ts":"2022-10-14T09:32:20Z","caller":"cmd/main.go:48","msg":"Constellation Node Join Service","version":"2.1.0","cloudProvider":"qemu"}
{"level":"INFO","ts":"2022-10-14T09:32:20Z","logger":"validator","caller":"watcher/validator.go:96","msg":"Updating expected measurements"}
{"level":"INFO","ts":"2022-10-14T09:32:20Z","logger":"server","caller":"server/server.go:73","msg":"Starting join service on [::]:9090"}
{"level":"INFO","ts":"2022-10-14T09:32:20Z","caller":"cmd/main.go:103","msg":"starting file watcher for measurements file /var/config/measurements"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:86","msg":"IssueJoinTicket called","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:88","msg":"Requesting measurement secret","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:41","msg":"Connecting to KMS at kms.kube-system:9000","keyID":"measurementSecret","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:48","msg":"Requesting data key","keyID":"measurementSecret","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:61","msg":"Data key request successful","keyID":"measurementSecret","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:95","msg":"Requesting disk encryption key","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:41","msg":"Connecting to KMS at kms.kube-system:9000","keyID":"0f87c61f-31e7-466d-be22-e7300e7d9e76","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:48","msg":"Requesting data key","keyID":"0f87c61f-31e7-466d-be22-e7300e7d9e76","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kms","caller":"kms/kms.go:61","msg":"Data key request successful","keyID":"0f87c61f-31e7-466d-be22-e7300e7d9e76","endpoint":"kms.kube-system:9000"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:102","msg":"Creating Kubernetes join token","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kubeadm","caller":"kubeadm/kubeadm.go:63","msg":"Generating new random bootstrap token"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kubeadm","caller":"kubeadm/kubeadm.go:81","msg":"Creating bootstrap token in Kubernetes"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kubeadm","caller":"kubeadm/kubeadm.go:87","msg":"Preparing join token for new node"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"kubeadm","caller":"kubeadm/kubeadm.go:109","msg":"Join token creation successful"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:109","msg":"Querying K8sVersion ConfigMap","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:115","msg":"Creating signed kubelet certificate","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"certificateAuthority","caller":"kubernetesca/kubernetesca.go:84","msg":"Creating kubelet certificate"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server","caller":"server/server.go:138","msg":"IssueJoinTicket successful","peerAddress":"10.42.2.100:59988"}
{"level":"INFO","ts":"2022-10-14T09:32:21Z","logger":"server.gRPC","caller":"zap/server_interceptors.go:39","msg":"finished unary call with code OK","grpc.start_time":"2022-10-14T09:32:21Z","grpc.request.deadline":"2022-10-14T09:32:51Z","system":"grpc","span.kind":"server","grpc.service":"join.API","grpc.method":"IssueJoinTicket","peer.address":"10.42.2.100:59988","grpc.code":"OK","grpc.time_ms":27.715}
```

Once the worker node has joined your cluster, it may take a couple of minutes for all resources to be available.
You can check on the state of your cluster by running the following:

```bash
kubectl get nodes
```

If your cluster is running as expected the output should look like the following:

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
