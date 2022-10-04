# Manage SSH keys

Constellation allows you to create UNIX users that can connect to both control-plane and worker nodes over SSH. As the system partitions are read-only, users need to be re-created upon each restart of a node. This is automated by the *Access Manager*.

On cluster initialization, users defined in the `ssh-users` section of the Constellation configuration file are created and stored in the `ssh-users` ConfigMap in the `kube-system` namespace. For a running cluster, you can add or remove users by modifying the ConfigMap and restarting a node.

## Access Manager
The Access Manager supports all OpenSSH key types. These are RSA, ECDSA (using the `nistp256`, `nistp384`, `nistp521` curves) and Ed25519.

:::note
All users are automatically created with `sudo` capabilities.
:::

The Access Manager is deployed as a DaemonSet called `constellation-access-manager`, running as an `initContainer` and afterward running a `pause` container to avoid automatic restarts. While technically killing the Pod and letting it restart works for the (re-)creation of users, it doesn't automatically remove users. Thus, a node restart is required after making changes to the ConfigMap.

When a user is deleted from the ConfigMap, it won't be re-created after the next restart of a node. The home directories of the affected users will be moved to `/var/evicted`.

You can update the ConfigMap by:
```bash
kubectl edit configmap -n kube-system ssh-users
```

Or alternatively, by modifying and re-applying it with the definition listed in the examples.

## Examples
You can add a user `myuser` in `constellation-config.yaml` like this:

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

You can add users by adding `data` entries:

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
