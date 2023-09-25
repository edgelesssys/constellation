terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.67.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
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
  uid            = random_id.uid.hex
  name           = "${var.name}-${local.uid}"
  initSecretHash = random_password.initSecret.bcrypt_hash
  tags = {
    constellation-uid = local.uid,
  }
  ports_node_range      = "30000-32767"
  ports_kubernetes      = "6443"
  ports_bootstrapper    = "9000"
  ports_konnectivity    = "8132"
  ports_verify          = "30081"
  ports_recovery        = "9999"
  ports_join            = "30090"
  ports_debugd          = "4000"
  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  cidr_vpc_subnet_pods  = "10.10.0.0/16"
  // wildcard_lb_dns_name is the DNS name of the load balancer with a wildcard for the name.
  // example: given "name-1234567890.location.cloudapp.azure.com" it will return "*.location.cloudapp.azure.com"
  wildcard_lb_dns_name = replace(data.azurerm_public_ip.loadbalancer_ip.fqdn, "/^[^.]*\\./", "*.")
  // deduce from format (subscriptions)/$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME"
  // move from the right as to ignore the optional prefixes
  uai_resource_group = element(split("/", var.user_assigned_identity), length(split("/", var.user_assigned_identity)) - 5)
  // deduce as above
  uai_name = element(split("/", var.user_assigned_identity), length(split("/", var.user_assigned_identity)) - 1)
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

  lifecycle {
    # Attestation policies will be set automatically upon creation, even if not specified in the resource,
    # while they aren't being incorporated into the Terraform state correctly.
    # To prevent them from being set to null when applying an upgrade, ignore the changes until the issue
    # is resolved by Azure.
    # Related issue: https://github.com/hashicorp/terraform-provider-azurerm/issues/21998
    ignore_changes = [open_enclave_policy_base64, sgx_enclave_policy_base64, tpm_policy_base64, sev_snp_policy_base64]
  }
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
  domain_name_label   = local.name
  resource_group_name = var.resource_group
  location            = var.location
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.tags

  lifecycle {
    ignore_changes = [name]
  }
}

// Reads data from the resource of the same name.
// Used to wait to the actual resource to become ready, before using data from that resource.
// Property "fqdn" only becomes available on azurerm_public_ip resources once domain_name_label is set.
// Since we are setting domain_name_label starting with 2.10 we need to migrate
// resources for clusters created before 2.9. In those cases we need to wait until loadbalancer_ip has
// been updated before reading from it.
data "azurerm_public_ip" "loadbalancer_ip" {
  name                = "${local.name}-lb"
  resource_group_name = var.resource_group
  depends_on          = [azurerm_public_ip.loadbalancer_ip]
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
    {
      name     = "join",
      port     = local.ports_join,
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
      { name = "join", priority = 104, dest_port_range = local.ports_recovery },
      { name = "recovery", priority = 105, dest_port_range = local.ports_join },
      var.debug ? [{ name = "debugd", priority = 106, dest_port_range = local.ports_debugd }] : [],
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

module "scale_set_group" {
  source          = "./modules/scale_set"
  for_each        = var.node_groups
  base_name       = local.name
  node_group_name = each.key
  role            = each.value.role
  zones           = each.value.zones
  tags = merge(
    local.tags,
    { constellation-init-secret-hash = local.initSecretHash },
    { constellation-maa-url = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : "" },
  )

  initial_count             = each.value.initial_count
  state_disk_size           = each.value.disk_size
  state_disk_type           = each.value.disk_type
  location                  = var.location
  instance_type             = each.value.instance_type
  confidential_vm           = var.confidential_vm
  secure_boot               = var.secure_boot
  resource_group            = var.resource_group
  user_assigned_identity    = var.user_assigned_identity
  image_id                  = var.image_id
  network_security_group_id = azurerm_network_security_group.security_group.id
  subnet_id                 = azurerm_subnet.node_subnet.id
  backend_address_pool_ids = each.value.role == "control-plane" ? [
    azurerm_lb_backend_address_pool.all.id,
    module.loadbalancer_backend_control_plane.backendpool_id
    ] : [
    azurerm_lb_backend_address_pool.all.id,
    module.loadbalancer_backend_worker.backendpool_id
  ]
}

data "azurerm_subscription" "current" {
}

data "azurerm_user_assigned_identity" "uaid" {
  name                = local.uai_name
  resource_group_name = local.uai_resource_group
}

moved {
  from = module.scale_set_control_plane
  to   = module.scale_set_group["control_plane_default"]
}

moved {
  from = module.scale_set_worker
  to   = module.scale_set_group["worker_default"]
}
