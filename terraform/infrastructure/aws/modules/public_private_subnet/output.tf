output "private_subnet_id" {
  value = {
    for az in data.aws_availability_zone.all :
    az.name => aws_subnet.private[az.name].id
  }
}

output "public_subnet_id" {
  value = {
    for az in data.aws_availability_zone.all :
    az.name => aws_subnet.public[az.name].id
  }
}

# all_zones is a list of all availability zones in the region
# it also contains zones that are not currently used by node groups (but might be in the future)
output "all_zones" {
  value = distinct(sort([for az in data.aws_availability_zone.all : az.name]))
}
