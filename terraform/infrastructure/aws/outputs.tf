output "out_of_cluster_endpoint" {
  value       = local.out_of_cluster_endpoint
  description = "External endpoint for the Kubernetes API server. Only varies from the `in_cluster_endpoint` when using an internal load balancer."
}

output "in_cluster_endpoint" {
  value       = local.in_cluster_endpoint
  description = "Internal endpoint for the Kubernetes API server."
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
  description = "List of additional Subject Alternative Names (SANs) for the API server certificate."
}

output "uid" {
  value       = local.uid
  description = "Unique Identifier (UID) of the cluster."
}

output "init_secret" {
  value       = random_password.init_secret.result
  sensitive   = true
  description = "The init secret for the cluster."
}

output "name" {
  value       = local.name
  description = "Unique name of the Constellation cluster, comprised by name and UID."
}

output "ip_cidr_node" {
  value       = local.cidr_vpc_subnet_nodes
  description = "CIDR block of the node network."
}
