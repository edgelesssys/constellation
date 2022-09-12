# Horizontal Pod Autoscaling
This example demonstrates Constellation's autoscaling capabilities by utilizing a slightly adapted version of the Kubernetes [HorizontalPodAutoscaler Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/). During the following steps, Constellation will spawn new VMs on demand, verify them, add them to the cluster, and delete them again when the load has settled down.

## Requirements
The cluster needs to be initialized with Kubernetes 1.23 or higher. In addition, autoscaling must be enabled to trigger Constellation to assign new nodes dynamically.

Just for this example specifically, the cluster should have as few worker nodes in the beginning as possible. We recommend starting with a small cluster with only *one* low-powered node for the control plane node and *one* low-powered worker node. 

:::info
We tested the example using instances of types `Standard_DC4as_v5` on Azure and `n2d-standard-4` on GCP.
:::

## Setup

1. Install the Kubernetes Metrics Server:
    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    ```

2. Deploy the HPA example server that's supposed to be scaled under load.

    This is almost the same as the example which can be found in the official Kubernetes HPA walkthrough, with the only difference being increased CPU limits and requests to facilitate the triggering of node scaling events.
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: php-apache
    spec:
      selector:
        matchLabels:
          run: php-apache
      replicas: 1
      template:
        metadata:
          labels:
            run: php-apache
        spec:
          containers:
          - name: php-apache
            image: registry.k8s.io/hpa-example
            ports:
            - containerPort: 80
            resources:
              limits:
                cpu: 900m
              requests:
                cpu: 600m
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: php-apache
      labels:
        run: php-apache
    spec:
      ports:
      - port: 80
      selector:
        run: php-apache
    EOF
    ```
3. Create a HorizontalPodAutoscaler.
    It's recommended to set an average CPU utilization across all Pods of 20% with the above server CPU limits and requests to see one additional worker nodes being created later. Note that the CPU utilization used here isn't 1:1 the host CPU utilization, but rather the requested CPU capacities (20% of 600 milli-cores CPU across all Pods). Take a look at the [original tutorial](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#create-horizontal-pod-autoscaler) for more information on the HPA configuration.
    ```bash
    kubectl autoscale deployment php-apache --cpu-percent=20 --min=1 --max=10
    ```
4. Create a Pod which generates load onto the server:
    ```bash
    kubectl run -i --tty load-generator --rm --image=busybox:1.28 --restart=Never -- /bin/sh -c "while true; do wget -q -O- http://php-apache; done"
    ```
5. Wait for a few minutes until new nodes are added to the cluster. See below for how to monitor the state of the HorizontalPodAutoscaler, the list of nodes and the behavior of the autoscaler.
6. To kill the load generator, press CTRL+C and run:
    ```bash
    kubectl delete pod load-generator
    ```
7. The cluster-autoscaler checks every few minutes if nodes are underutilized and can be removed from the cluster. It will taint such candidates for removal and wait additional 10 minutes before the nodes are eventually removed and deallocated. The whole process can take ~20 minutes in total.

## Monitoring
:::tip

For better observability, run the listed commands in different tabs in your terminal.

:::

You can watch the status of the HorizontalPodAutoscaler with the current CPU, the target CPU limit, and the number of replicas created with:
```bash
kubectl get hpa php-apache --watch
```
From time to time compare the list of nodes to check the behavior of the autoscaler:
```bash
kubectl get nodes
```
For deeper insights, take a look at the logs of the autoscaler Pod which contains more details about the scaling decision process:
```bash
kubectl logs -f deployment/constellation-cluster-autoscaler -n kube-system
```
