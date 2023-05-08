terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.55.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.4.3"
    }
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
}

locals {
  uid                   = random_id.uid.hex
  name                  = "${var.name}-${local.uid}"
  initSecretHash        = random_password.initSecret.bcrypt_hash
  tags                  = { constellation-uid = local.uid }
  ports_node_range      = "30000-32767"
  ports_kubernetes      = "6443"
  ports_bootstrapper    = "9000"
  ports_konnectivity    = "8132"
  ports_verify          = "30081"
  ports_recovery        = "9999"
  ports_debugd          = "4000"
  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  cidr_vpc_subnet_pods  = "10.10.0.0/16"
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "initSecret" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "azurerm_attestation_provider" "attestation_provider" {
  count = var.create_maa ? 1 : 0
  # name must be between 3 and 24 characters in length and use numbers and lower-case letters only.
  name                = format("constell%s", local.uid)
  resource_group_name = var.resource_group
  location            = var.location
}

resource "azurerm_application_insights" "insights" {
  name                = local.name
  location            = var.location
  resource_group_name = var.resource_group
  application_type    = "other"
  tags                = local.tags
}

resource "azurerm_public_ip" "loadbalancer_ip" {
  name                = "${local.name}-lb"
  resource_group_name = var.resource_group
  location            = var.location
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.tags
}

resource "azurerm_public_ip" "nat_gateway_ip" {
  name                = "${local.name}-nat"
  resource_group_name = var.resource_group
  location            = var.location
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.tags
}

resource "azurerm_nat_gateway" "gateway" {
  name                    = local.name
  location                = var.location
  resource_group_name     = var.resource_group
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
}

resource "azurerm_subnet_nat_gateway_association" "example" {
  nat_gateway_id = azurerm_nat_gateway.gateway.id
  subnet_id      = azurerm_subnet.node_subnet.id
}

resource "azurerm_nat_gateway_public_ip_association" "example" {
  nat_gateway_id       = azurerm_nat_gateway.gateway.id
  public_ip_address_id = azurerm_public_ip.nat_gateway_ip.id
}

resource "azurerm_lb" "loadbalancer" {
  name                = local.name
  location            = var.location
  resource_group_name = var.resource_group
  sku                 = "Standard"
  tags                = local.tags

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.loadbalancer_ip.id
  }
}

module "loadbalancer_backend_control_plane" {
  source = "./modules/load_balancer_backend"

  name            = "${local.name}-control-plane"
  loadbalancer_id = azurerm_lb.loadbalancer.id
  ports = flatten([
    {
      name     = "bootstrapper",
      port     = local.ports_bootstrapper,
      protocol = "Tcp",
      path     = null
    },
    {
      name     = "kubernetes",
      port     = local.ports_kubernetes,
      protocol = "Https",
      path     = "/readyz"
    },
    {
      name     = "konnectivity",
      port     = local.ports_konnectivity,
      protocol = "Tcp",
      path     = null
    },
    {
      name     = "verify",
      port     = local.ports_verify,
      protocol = "Tcp",
      path     = null
    },
    {
      name     = "recovery",
      port     = local.ports_recovery,
      protocol = "Tcp",
      path     = null
    },
    var.debug ? [{
      name     = "debugd",
      port     = local.ports_debugd,
      protocol = "Tcp",
      path     = null
    }] : [],
  ])
}

module "loadbalancer_backend_worker" {
  source = "./modules/load_balancer_backend"

  name            = "${local.name}-worker"
  loadbalancer_id = azurerm_lb.loadbalancer.id
  ports           = []
}

resource "azurerm_lb_backend_address_pool" "all" {
  loadbalancer_id = azurerm_lb.loadbalancer.id
  name            = "${var.name}-all"
}

resource "azurerm_virtual_network" "network" {
  name                = local.name
  resource_group_name = var.resource_group
  location            = var.location
  address_space       = ["10.0.0.0/8"]
  tags                = local.tags
}

resource "azurerm_subnet" "node_subnet" {
  name                 = "${local.name}-node"
  resource_group_name  = var.resource_group
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = ["10.9.0.0/16"]
}

resource "azurerm_network_security_group" "security_group" {
  name                = local.name
  location            = var.location
  resource_group_name = var.resource_group
  tags                = local.tags

  dynamic "security_rule" {
    for_each = flatten([
      { name = "noderange", priority = 100, dest_port_range = local.ports_node_range },
      { name = "kubernetes", priority = 101, dest_port_range = local.ports_kubernetes },
      { name = "bootstrapper", priority = 102, dest_port_range = local.ports_bootstrapper },
      { name = "konnectivity", priority = 103, dest_port_range = local.ports_konnectivity },
      { name = "recovery", priority = 104, dest_port_range = local.ports_recovery },
      var.debug ? [{ name = "debugd", priority = 105, dest_port_range = local.ports_debugd }] : [],
    ])
    content {
      name                       = security_rule.value.name
      priority                   = security_rule.value.priority
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = security_rule.value.dest_port_range
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    }
  }
}

module "scale_set_control_plane" {
  source = "./modules/scale_set"

  name            = "${local.name}-control-plane"
  instance_count  = var.control_plane_count
  state_disk_size = var.state_disk_size
  state_disk_type = var.state_disk_type
  resource_group  = var.resource_group
  location        = var.location
  instance_type   = var.instance_type
  confidential_vm = var.confidential_vm
  secure_boot     = var.secure_boot
  tags = merge(
    local.tags,
    { constellation-role = "control-plane" },
    { constellation-init-secret-hash = local.initSecretHash },
    { constellation-maa-url = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : "" },
  )
  image_id                  = var.image_id
  user_assigned_identity    = var.user_assigned_identity
  network_security_group_id = azurerm_network_security_group.security_group.id
  subnet_id                 = azurerm_subnet.node_subnet.id
  backend_address_pool_ids = [
    azurerm_lb_backend_address_pool.all.id,
    module.loadbalancer_backend_control_plane.backendpool_id
  ]
}

module "scale_set_worker" {
  source = "./modules/scale_set"

  name            = "${local.name}-worker"
  instance_count  = var.worker_count
  state_disk_size = var.state_disk_size
  state_disk_type = var.state_disk_type
  resource_group  = var.resource_group
  location        = var.location
  instance_type   = var.instance_type
  confidential_vm = var.confidential_vm
  secure_boot     = var.secure_boot
  tags = merge(
    local.tags,
    { constellation-role = "worker" },
    { constellation-init-secret-hash = local.initSecretHash },
    { constellation-maa-url = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : "" },
  )
  image_id                  = var.image_id
  user_assigned_identity    = var.user_assigned_identity
  network_security_group_id = azurerm_network_security_group.security_group.id
  subnet_id                 = azurerm_subnet.node_subnet.id
  backend_address_pool_ids = [
    azurerm_lb_backend_address_pool.all.id,
    module.loadbalancer_backend_worker.backendpool_id,
  ]
}
