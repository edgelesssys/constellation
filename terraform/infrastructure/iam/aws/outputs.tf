output "control_plane_instance_profile_name" {
  value       = aws_iam_instance_profile.control_plane_instance_profile.name
  description = "Name of the control plane's instance profile."
}

output "worker_nodes_instance_profile_name" {
  value       = aws_iam_instance_profile.worker_node_instance_profile.name
  description = "Name of the worker nodes' instance profile"
}
