output "out_of_cluster_endpoint" {
  value = module.node_group["control_plane_default"].instance_ips[0]
}

output "in_cluster_endpoint" {
  value = module.node_group["control_plane_default"].instance_ips[0]
}

output "extra_api_server_cert_sans" {
  value = sort(concat([module.node_group["control_plane_default"].instance_ips[0]], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "uid" {
  value = "qemu" // placeholder
}

output "init_secret" {
  value     = random_password.init_secret.result
  sensitive = true
}

output "validate_constellation_kernel" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_kernel) > 0
    error_message = "constellation_kernel must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
}

output "validate_constellation_initrd" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_initrd) > 0
    error_message = "constellation_initrd must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
}

output "validate_constellation_cmdline" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_cmdline) > 0
    error_message = "constellation_cmdline must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
}

output "name" {
  value = "${var.name}-qemu" // placeholder, as per "uid" output
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
}
