output "ips" {
  value       = [for instance in openstack_compute_instance_v2.instance_group_member : instance.access_ip_v4]
  description = "Public IP addresses of the instances."
}

output "instance_ids" {
  value       = openstack_compute_instance_v2.instance_group_member.*.id
  description = "IDs of the instances."
}

output "port_ids" {
  value       = openstack_networking_port_v2.port.*.id
  description = "IDs of ports of the instances."
}
