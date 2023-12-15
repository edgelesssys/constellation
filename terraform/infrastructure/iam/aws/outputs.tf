output "iam_instance_profile_name_control_plane" {
  value       = aws_iam_instance_profile.control_plane_instance_profile.name
  description = "Name of the control plane's instance profile."
}

output "iam_instance_profile_name_worker_nodes" {
  value       = aws_iam_instance_profile.worker_node_instance_profile.name
  description = "Name of the worker nodes' instance profile"
}
