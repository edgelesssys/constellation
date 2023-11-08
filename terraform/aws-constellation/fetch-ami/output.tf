output "ami" {
  description = "The fetched AMI."
  value       = data.local_file.ami.content
}
