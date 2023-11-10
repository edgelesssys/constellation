output "image" {
  description = "The fetched image data."
  value       = data.local_file.image.content
}
