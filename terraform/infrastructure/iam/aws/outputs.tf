output "control_plane_instance_profile" {
  value = aws_iam_instance_profile.control_plane_instance_profile.name
  description = "The name of the instance profile for the control plane."
}

output "worker_nodes_instance_profile" {
  value = aws_iam_instance_profile.worker_node_instance_profile.name
  description = "The name of the instance profile for the worker nodes."
}
