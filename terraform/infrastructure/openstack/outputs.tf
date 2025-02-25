# Outputs common to all CSPs

output "out_of_cluster_endpoint" {
  value       = openstack_networking_floatingip_v2.public_ip.address
  description = "External endpoint for the Kubernetes API server. Only varies from the `in_cluster_endpoint` when using an internal load balancer."
}

output "in_cluster_endpoint" {
  value       = openstack_networking_floatingip_v2.public_ip.address
  description = "Internal endpoint for the Kubernetes API server."
}

output "api_server_cert_sans" {
  value       = sort(concat([openstack_networking_floatingip_v2.public_ip.address], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
  description = "List of Subject Alternative Names (SANs) for the API server certificate."
}

output "uid" {
  value       = local.uid
  description = "Unique Identifier (UID) of the cluster."
}

output "init_secret" {
  value       = random_password.init_secret.result
  sensitive   = true
  description = "Initialization secret to authenticate the bootstrapping node."
}

output "name" {
  value       = local.name
  description = "Unique name of the Constellation cluster, comprised by name and UID."
}

output "ip_cidr_node" {
  value       = local.cidr_vpc_subnet_nodes
  description = "CIDR block of the node network."
}

output "loadbalancer_address" {
  value       = openstack_networking_floatingip_v2.public_ip.address
  description = "Public loadbalancer address."
}

# OpenStack-specific outputs

output "network_id" {
  value       = openstack_networking_network_v2.vpc_network.id
  description = "The OpenStack network id the cluster is deployed in."
}

output "lb_subnetwork_id" {
  value       = openstack_networking_subnet_v2.lb_subnetwork.id
  description = "The OpenStack subnetwork id lbs are deployed in."
}
