output "instance_ips" {
  value = flatten(libvirt_domain.instance_group[*].network_interface[*].addresses[*])
}
