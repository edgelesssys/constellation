output "image" {
  description = "The resolved image ID of the CSP."
  value       = data.local_file.image.content
}
