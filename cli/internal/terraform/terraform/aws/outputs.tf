output "out_of_cluster_endpoint" {
  value = local.out_of_cluster_endpoint
}

output "in_cluster_endpoint" {
  value = local.in_cluster_endpoint
}
output "api_server_cert_sans" {
  value = sort(
    distinct(
      concat(
        [
          local.in_cluster_endpoint,
          local.out_of_cluster_endpoint,
        ],
        var.custom_endpoint == "" ? [] : [var.custom_endpoint],
      )
    )
  )
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}

output "name" {
  value = local.name
}
