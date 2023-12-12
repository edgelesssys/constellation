variable "name" {
  type        = string
  description = "Name of the Constellation cluster."
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC."
}

variable "zone" {
  type        = string
  description = "Main availability zone. Only used for legacy reasons."
}

variable "zones" {
  type        = list(string)
  description = "Availability zones."
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
  description = "Zags to add to the resource."
}
