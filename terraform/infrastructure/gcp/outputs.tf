output "out_of_cluster_endpoint" {
  value = local.out_of_cluster_endpoint
  description = "The external endpoint for the Kubernetes API server. This only varies from the `in_cluster_endpoint` when using an internal load balancer setup."
}

output "in_cluster_endpoint" {
  value = local.in_cluster_endpoint
  description = "The internal endpoint for the Kubernetes API server. This only varies from the `in_cluster_endpoint` when using an internal load balancer setup."
}
output "api_server_cert_sans" {
  value = sort(
    distinct(
      concat(
        [
          local.in_cluster_endpoint,
          local.out_of_cluster_endpoint,
        ],
        var.custom_endpoint == "" ? [] : [var.custom_endpoint],
      )
    )
  )
  description = "List of Subject Alternative Names (SANs) for the API server certificate."
}

output "uid" {
  value = local.uid
  description = "The UID of the cluster."
}

output "init_secret" {
  value     = random_password.initSecret.result
  sensitive = true
  description = "The init secret for the cluster."
}

output "project" {
  value = var.project
  description = "The GCP project to deploy the cluster in."
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
}

output "ip_cidr_pods" {
  value = local.cidr_vpc_subnet_pods
}

output "name" {
  value = local.name
}
