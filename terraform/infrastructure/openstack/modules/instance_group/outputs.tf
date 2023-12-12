output "ips" {
  value = openstack_compute_instance_v2.instance_group_member.*.access_ip_v4
  description = "Public IP addresses of the instances."
}

output "instance_ids" {
  value = openstack_compute_instance_v2.instance_group_member.*.id
  description = "IDs of the instances."
}
