# Mini Constellation

With `constellation mini`, users can deploy and test Constellation locally without the need for a cloud subscription.

The command uses virtualization to create a local cluster with one control-plane and one worker node.

:::info

Since mini Constellation is running on your local system, please note that common cloud features, such as load-balancing,
attaching persistent storage, or autoscaling, are unavailable.

:::

## Prerequisites

* [Constellation CLI](./install.md#install-the-constellation-cli)
* A x86-64 CPU with at least 4 cores
  * Recommended are 6 cores or more
* Hardware virtualization enabled in the BIOS/UEFI (often referred to as Intel VT-x or AMD-V/SVM)
* At least 4 GB RAM
  * Recommend are 6 GB or more
* 20 GB of free disk space
* a Linux operating system
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

Setting up your mini Constellation cluster is as easy as running the following command:

```bash
constellation mini up
```

This will configure your current directory as the working directory for Constellation.
All `constellation` commands concerning this cluster need to be issued from this directory.

The command will create your cluster and initialize it. Depending on your system, this may take up to 10 minutes.
The output should look like the following:

```shell
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

You can now configure `kubectl` to connect to your local Constellation cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
```

It may take a couple of minutes for all cluster resources to be available.
You can check on the state of your cluster by running the following:

```bash
kubectl get nodes
```

If your cluster is running as expected the output should look like the following:

```shell
$ kubectl get nodes
NAME              STATUS   ROLES                  AGE     VERSION
control-plane-0   Ready    control-plane,master   2m59s   v1.23.9
worker-0          Ready    <none>                 32s     v1.23.9
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

```shell
constellation mini down
```

This will destroy your cluster and clean up the your working directory.
The VM image and cluster configuration file (`constellation-conf.yaml`) will be left behind and may be reused to create new clusters.

## Troubleshooting

### VMs have no internet access

`iptables` rules may prevent your VMs form properly accessing the internet.
Make sure your rules are'nt dropping forwarded packages.

List your rules:

```shell
sudo iptables -S
```

The output may look similar to the following:

```shell
-P INPUT ACCEPT
-P FORWARD DROP
-P OUTPUT ACCEPT
-N DOCKER
-N DOCKER-ISOLATION-STAGE-1
-N DOCKER-ISOLATION-STAGE-2
-N DOCKER-USER
```

If your `FORWARD` chain is set to `DROP`, you will need to update your rules:

```shell
sudo iptables -P FORWARD ACCEPT
```
