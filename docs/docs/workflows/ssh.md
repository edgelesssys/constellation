# Manage SSH Keys

Constellation gives you the capability to create UNIX users which can connect to the cluster nodes over SSH, allowing you to access both control-plane as well as worker nodes. While the data partition is persistent, the system partition is read-only, meaning that users need to be re-created upon each restart of a node. This is where the Access Manager comes into effect, ensuring the automatic (re-)creation of all users whenever a node is restarted.

During the initial creation of the cluster, all users defined in the `ssh-users` section of the Constellation configuration (see the [reference section](../reference/config.md) for details) are automatically created during the initialization process.

For persistence, they're transferred into a ConfigMap called `ssh-users`, residing in the `kube-system` namespace. When no users are initially defined, the ConfigMap will still be created with no entries. After the initial definition in the Constellation configuration, users can be added and removed by modifying the entries of the ConfigMap and performing a restart of a node.

## Access Manager
The Access Manager doesn't restrict users on the use of certain key formats, meaning that all underlying formats the OpenSSH server supports are accepted. These are RSA, ECDSA (using the `nistp256`, `nistp384`, `nistp521` curves) and Ed25519. No validation is performed on the side of the Access Manager too, passing them directly to the authorized key lists as defined.

Note that all users are automatically created with `sudo` capabilities, so make sure no one unintended has permissions to modify the `ssh-users` ConfigMap.

The Access Manager is deployed as a DaemonSet called `constellation-access-manager`, running as an `initContainer` and afterward running a `pause` container to avoid automatic restarts. While technically killing the Pod and letting it restart works for the (re-)creation of users, it doesn't automatically remove users. Therefore, a complete node restart is important to ensure the correct modification of users on the system and needs to be executed manually after making changes to the ConfigMap.

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
