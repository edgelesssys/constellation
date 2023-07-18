# Expose services

## Internet-facing LB service on AWS

To expose your application service externally you might want to use a Kubernetes Service of type `LoadBalancer`. On AWS, load-balancing is achieved through the [AWS Load Balancing Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller) as in the managed EKS.

Since recent versions, the controller deploy an internal LB by default requiring to set an annotation `service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing` in order to have an internet-facing LB. For more details, see the [official docs](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/service/nlb/).

:::caution
Before terminating the cluster, all LB backed services should be deleted, so that the controller can cleanup the related resources.
:::
