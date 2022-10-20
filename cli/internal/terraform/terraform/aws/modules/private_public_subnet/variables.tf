variable "name" {
  type        = string
  description = "Name of your Constellation, which is used as a prefix for tags."
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC."
}

variable "zone" {
  type        = string
  description = "Availability zone."
}

variable "cidr_vpc_subnet_nodes" {
  type        = string
  description = "CIDR block for the subnet that will contain the nodes."
}

variable "cidr_vpc_subnet_internet" {
  type        = string
  description = "CIDR block for the subnet that contains resources reachable from the Internet."
}

variable "tags" {
  type        = map(string)
  description = "The tags to add to the resource."
}
