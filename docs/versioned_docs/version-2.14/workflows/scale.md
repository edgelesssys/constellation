# Scale your cluster

Constellation provides all features of a Kubernetes cluster including scaling and autoscaling.

## Worker node scaling

### Autoscaling

Constellation comes with autoscaling disabled by default. To enable autoscaling, find the scaling group of
worker nodes:

```bash
kubectl get scalinggroups -o json | yq '.items | .[] | select(.spec.role == "Worker") | [{"name": .metadata.name, "nodeGoupName": .spec.nodeGroupName}]'
```

This will output a list of scaling groups with the corresponding cloud provider name (`name`) and the cloud provider agnostic name of the node group (`nodeGroupName`).

Then, patch the `autoscaling` field of the scaling group resource with the desired `name` to `true`:

```bash
# Replace <name> with the name of the scaling group you want to enable autoscaling for
worker_group=<name>
kubectl patch scalinggroups $worker_group --patch '{"spec":{"autoscaling": true}}' --type='merge'
kubectl get scalinggroup $worker_group -o jsonpath='{.spec}' | yq -P
```

The cluster autoscaler now automatically provisions additional worker nodes so that all pods have a place to run.
You can configure the minimum and maximum number of worker nodes in the scaling group by patching the `min` or
`max` fields of the scaling group resource:

```bash
kubectl patch scalinggroups $worker_group --patch '{"spec":{"max": 5}}' --type='merge'
kubectl get scalinggroup $worker_group -o jsonpath='{.spec}' | yq -P
```

The cluster autoscaler will now never provision more than 5 worker nodes.

If you want to see the autoscaling in action, try to add a deployment with a lot of replicas, like the
following Nginx deployment. The number of replicas needed to trigger the autoscaling depends on the size of
and count of your worker nodes. Wait for the rollout of the deployment to finish and compare the number of
worker nodes before and after the deployment:

```bash
kubectl create deployment nginx --image=nginx --replicas 150
kubectl -n kube-system get nodes
kubectl rollout status deployment nginx
kubectl -n kube-system get nodes
```

### Manual scaling

Alternatively, you can manually scale your cluster up or down:

<tabs groupId="csp">
<tabItem value="azure" label="Azure">

1. Find your Constellation resource group.
2. Select the `scale-set-workers`.
3. Go to **settings** and **scaling**.
4. Set the new **instance count** and **save**.

</tabItem>
<tabItem value="gcp" label="GCP">

1. In Compute Engine go to [Instance Groups](https://console.cloud.google.com/compute/instanceGroups/).
2. **Edit** the **worker** instance group.
3. Set the new **number of instances** and **save**.

</tabItem>
<tabItem value="aws" label="AWS">

1. Go to Auto Scaling Groups and select the worker ASG to scale up.
2. Click **Edit**
3. Set the new (increased) **Desired capacity** and **Update**.

</tabItem>
</tabs>

## Control-plane node scaling

Control-plane nodes can **only be scaled manually and only scaled up**!

To increase the number of control-plane nodes, follow these steps:

<tabs groupId="csp">

<tabItem value="azure" label="Azure">

1. Find your Constellation resource group.
2. Select the `scale-set-controlplanes`.
3. Go to **settings** and **scaling**.
4. Set the new (increased) **instance count** and **save**.

</tabItem>
<tabItem value="gcp" label="GCP">

1. In Compute Engine go to [Instance Groups](https://console.cloud.google.com/compute/instanceGroups/).
2. **Edit** the **control-plane** instance group.
3. Set the new (increased) **number of instances** and **save**.

</tabItem>
<tabItem value="aws" label="AWS">

1. Go to Auto Scaling Groups and select the control-plane ASG to scale up.
2. Click **Edit**
3. Set the new (increased) **Desired capacity** and **Update**.

</tabItem>
</tabs>

If you scale down the number of control-planes nodes, the removed nodes won't be able to exit the `etcd` cluster correctly. This will endanger the quorum that's required to run a stable Kubernetes control plane.
