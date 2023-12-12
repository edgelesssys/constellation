output "instance_group_url" {
  value = google_compute_instance_group_manager.instance_group_manager.instance_group
  description = "Full URL of the instance group."
}
