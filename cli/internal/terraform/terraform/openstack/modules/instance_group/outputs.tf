output "instance_group" {
  value = local.name
}

output "ips" {
  value = openstack_compute_instance_v2.instance_group_member.*.access_ip_v4
}

output "instance_ids" {
  value = openstack_compute_instance_v2.instance_group_member.*.id
}
