# Expose a service

Constellation integrates the native load balancers of each CSP. Therefore, to expose a service simply [create a service of type `LoadBalancer`](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer).

## Internet-facing LB service on AWS

To expose your application service externally you might want to use a Kubernetes Service of type `LoadBalancer`. On AWS, load-balancing is achieved through the [AWS Load Balancing Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller) as in the managed EKS.

Since recent versions, the controller deploy an internal LB by default requiring to set an annotation `service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing` to have an internet-facing LB. For more details, see the [official docs](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/service/nlb/).

For general information on LB with AWS see [Network load balancing on Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/network-load-balancing.html).

:::caution
Before terminating the cluster, all LB backed services should be deleted, so that the controller can cleanup the related resources.
:::

## Ingress on AWS

TODO(burgerdev): document
