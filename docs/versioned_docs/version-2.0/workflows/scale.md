# Scale your cluster

Constellation provides all features of a Kubernetes cluster including scaling and autoscaling.

## Worker node scaling

[During cluster initialization](create.md#the-init-step) you can choose to deploy the [cluster autoscaler](https://github.com/kubernetes/autoscaler). It automatically provisions additional worker nodes so that all pods have a place to run. Alternatively, you can choose to manually scale your cluster up or down:

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

1. Find your Constellation resource group.
2. Select the `scale-set-workers`.
3. Go to **settings** and **scaling**.
4. Set the new **instance count** and **save**.

</TabItem>
<TabItem value="gcp" label="GCP">

1. In Compute Engine go to [Instance Groups](https://console.cloud.google.com/compute/instanceGroups/).
2. **Edit** the **worker** instance group.
3. Set the new **number of instances** and **save**.

</TabItem>
</Tabs>

## Control-plane node scaling

Control-plane nodes can **only be scaled manually and only scaled up**!

To increase the number of control-plane nodes, follow these steps:

<Tabs groupId="csp">

<TabItem value="azure" label="Azure">

1. Find your Constellation resource group.
2. Select the `scale-set-controlplanes`.
3. Go to **settings** and **scaling**.
4. Set the new (increased) **instance count** and **save**.

</TabItem>
<TabItem value="gcp" label="GCP">

1. In Compute Engine go to [Instance Groups](https://console.cloud.google.com/compute/instanceGroups/).
2. **Edit** the **control-plane** instance group.
3. Set the new (increased) **number of instances** and **save**.

</TabItem>
</Tabs>

If you scale down the number of control-planes nodes, the removed nodes won't be able to exit the `etcd` cluster correctly. This will endanger the quorum that's required to run a stable Kubernetes control plane.
