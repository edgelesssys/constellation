# Manage SSH Keys

Constellation gives you the capability to create UNIX users which can connect to the cluster nodes over SSH, allowing you to access both control-plane and worker nodes. While the nodes' data partitions are persistent, the system partitions are read-only. Consequently, users need to be re-created upon each restart of a node. This is where the Access Manager comes into effect, ensuring the automatic (re-)creation of all users whenever a node is restarted.

During the initial creation of the cluster, all users defined in the `ssh-users` section of the Constellation [configuration file](../reference/config.md) are automatically created during the initialization process. For persistence, the users are stored in a ConfigMap called `ssh-users`, residing in the `kube-system` namespace. For a running cluster, users can be added and removed by modifying the entries of the ConfigMap and performing a restart of a node.

## Access Manager
The Access Manager supports all OpenSSH key types. These are RSA, ECDSA (using the `nistp256`, `nistp384`, `nistp521` curves) and Ed25519. 

:::note
All users are automatically created with `sudo` capabilities.
:::

The Access Manager is deployed as a DaemonSet called `constellation-access-manager`, running as an `initContainer` and afterward running a `pause` container to avoid automatic restarts. While technically killing the Pod and letting it restart works for the (re-)creation of users, it doesn't automatically remove users. Thus, a complete node restart is required after making changes to the ConfigMap.

When a user is deleted from the ConfigMap, it won't be re-created after the next restart of a node. The home directories of the affected users will be moved to `/var/evicted`, with the owner of each directory and its content being modified to `root`.

You can update the ConfigMap by:
```bash
kubectl edit configmap -n kube-system ssh-users
```

Or alternatively, by modifying and re-applying it with the definition listed in the examples.

## Examples
An example to create an user called `myuser` as part of the `constellation-config.yaml` looks like this:

```yaml
# Create SSH users on Constellation nodes upon the first initialization of the cluster.
sshUsers:
  myuser: "ssh-rsa AAAA...mgNJd9jc="
```

This user is then created upon the first initialization of the cluster, and translated into a ConfigMap as shown below:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ssh-users
  namespace: kube-system
data:
  myuser: "ssh-rsa AAAA...mgNJd9jc="
```

Entries can be added simply by adding `data` entries:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ssh-users
  namespace: kube-system
data:
  myuser: "ssh-rsa AAAA...mgNJd9jc="
  anotheruser: "ssh-ed25519 AAAA...CldH"
```

Similarly, removing any entries causes users to be evicted upon the next restart of the node.
