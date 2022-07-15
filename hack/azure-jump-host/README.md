# Creating a Jump Host for Azure

Constellation on Azure does not allow direct access to every node.
For debugging purposes, you can create a jump host that can be used to access the nodes in your cluster.

```shell-session
# execute the following command in your constellation workspace AFTER constellation create
"$(git rev-parse --show-toplevel)/hack/azure-jump-host/jump-host-create"
```
