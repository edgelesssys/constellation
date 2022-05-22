output "control_plane_ips" {
  value = module.control_plane.instance_ips
}

output "worker_ips" {
  value = module.worker.instance_ips
}
