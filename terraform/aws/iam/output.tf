output "control_plane_instance_profile" {
  value = aws_iam_instance_profile.control_plane_instance_profile.name
}

output "worker_nodes_instance_profile" {
  value = aws_iam_instance_profile.worker_node_instance_profile.name
}
