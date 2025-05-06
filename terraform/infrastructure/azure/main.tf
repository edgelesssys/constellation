terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.27.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
  subscription_id = var.subscription_id
  # This enables all resource providers.
  # In the future, we might want to use `resource_providers_to_register` to registers just the ones we need.
  resource_provider_registrations = "all"
}

locals {
  uid              = random_id.uid.hex
  name             = "${var.name}-${local.uid}"
  init_secret_hash = random_password.init_secret.bcrypt_hash
  tags = merge(
    var.additional_tags,
    { constellation-uid = local.uid }
  )
  ports_node_range      = "30000-32767"
  cidr_vpc_subnet_nodes = "10.9.0.0/16"
  ports = flatten([
    { name = "kubernetes", port = "6443", health_check_protocol = "Https", path = "/readyz", priority = 100 },
    { name = "bootstrapper", port = "9000", health_check_protocol = "Tcp", path = null, priority = 101 },
    { name = "verify", port = "30081", health_check_protocol = "Tcp", path = null, priority = 102 },
    { name = "recovery", port = "9999", health_check_protocol = "Tcp", path = null, priority = 104 },
    { name = "join", port = "30090", health_check_protocol = "Tcp", path = null, priority = 105 },
    var.debug ? [{ name = "debugd", port = "4000", health_check_protocol = "Tcp", path = null, priority = 106 }] : [],
    var.emergency_ssh ? [{ name = "ssh", port = "22", health_check_protocol = "Tcp", path = null, priority = 107 }] : [],
  ])
  // wildcard_lb_dns_name is the DNS name of the load balancer with a wildcard for the name.
  // example: given "name-1234567890.location.cloudapp.azure.com" it will return "*.location.cloudapp.azure.com"
  wildcard_lb_dns_name = var.internal_load_balancer ? "" : replace(data.azurerm_public_ip.loadbalancer_ip[0].fqdn, "/^[^.]*\\./", "*.")
  // deduce from format (subscriptions)/$ID/resourceGroups/$RG/providers/Microsoft.ManagedIdentity/userAssignedIdentities/$NAME"
  // move from the right as to ignore the optional prefixes
  uai_resource_group = element(split("/", var.user_assigned_identity), length(split("/", var.user_assigned_identity)) - 5)
  // deduce as above
  uai_name = element(split("/", var.user_assigned_identity), length(split("/", var.user_assigned_identity)) - 1)

  in_cluster_endpoint     = var.internal_load_balancer ? azurerm_lb.loadbalancer.frontend_ip_configuration[0].private_ip_address : azurerm_public_ip.loadbalancer_ip[0].ip_address
  out_of_cluster_endpoint = var.debug && var.internal_load_balancer ? module.jump_host[0].ip : local.in_cluster_endpoint
  revision                = 1
}

# A way to force replacement of resources if the provider does not want to replace them
# see: https://developer.hashicorp.com/terraform/language/resources/terraform-data#example-usage-data-for-replace_triggered_by
resource "terraform_data" "replacement" {
  input = local.revision
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "init_secret" {
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

  tags = local.tags
}

resource "azurerm_public_ip" "loadbalancer_ip" {
  count               = var.internal_load_balancer ? 0 : 1
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
  count               = var.internal_load_balancer ? 0 : 1
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
  tags                    = local.tags
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

  dynamic "frontend_ip_configuration" {
    for_each = var.internal_load_balancer ? [] : [1]
    content {
      name                 = "PublicIPAddress"
      public_ip_address_id = azurerm_public_ip.loadbalancer_ip[0].id
    }
  }

  dynamic "frontend_ip_configuration" {
    for_each = var.internal_load_balancer ? [1] : []
    content {
      name                          = "PrivateIPAddress"
      private_ip_address_allocation = "Dynamic"
      subnet_id                     = azurerm_subnet.loadbalancer_subnet[0].id
    }
  }
}

module "loadbalancer_backend_control_plane" {
  source = "./modules/load_balancer_backend"

  name                           = "${local.name}-control-plane"
  loadbalancer_id                = azurerm_lb.loadbalancer.id
  frontend_ip_configuration_name = azurerm_lb.loadbalancer.frontend_ip_configuration[0].name
  ports                          = local.ports
}

# We cannot delete them right away since we first need to to delete the dependency from the VMSS to this backend pool.
# TODO(@3u13r): Remove this resource after v2.18.0 has been released.
module "loadbalancer_backend_worker" {
  source = "./modules/load_balancer_backend"

  name                           = "${local.name}-worker"
  loadbalancer_id                = azurerm_lb.loadbalancer.id
  frontend_ip_configuration_name = azurerm_lb.loadbalancer.frontend_ip_configuration[0].name
  ports                          = []
}

# We cannot delete them right away since we first need to to delete the dependency from the VMSS to this backend pool.
# TODO(@3u13r): Remove this resource after v2.18.0 has been released.
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

resource "azurerm_subnet" "loadbalancer_subnet" {
  count                = var.internal_load_balancer ? 1 : 0
  name                 = "${local.name}-lb"
  resource_group_name  = var.resource_group
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = ["10.10.0.0/16"]
}

resource "azurerm_subnet" "node_subnet" {
  name                 = "${local.name}-node"
  resource_group_name  = var.resource_group
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = [local.cidr_vpc_subnet_nodes]
}

resource "azurerm_network_security_group" "security_group" {
  name                = local.name
  location            = var.location
  resource_group_name = var.resource_group
  tags                = local.tags
}

resource "azurerm_network_security_rule" "nsg_rule" {
  for_each = {
    for o in local.ports : o.name => o
  }
  # TODO(elchead): v2.20.0: remove name suffix and priority offset. Might need to add create_before_destroy to the NSG rule.
  name                        = "${each.value.name}-new"
  priority                    = each.value.priority + 10 # offset to not overlap with old rules
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = each.value.port
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = var.resource_group
  network_security_group_name = azurerm_network_security_group.security_group.name
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
    { constellation-init-secret-hash = local.init_secret_hash },
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
  backend_address_pool_ids  = each.value.role == "control-plane" ? [module.loadbalancer_backend_control_plane.backendpool_id] : []
  marketplace_image         = var.marketplace_image
}

module "jump_host" {
  count          = var.internal_load_balancer && var.debug ? 1 : 0
  source         = "./modules/jump_host"
  base_name      = local.name
  resource_group = var.resource_group
  location       = var.location
  subnet_id      = azurerm_subnet.loadbalancer_subnet[0].id
  ports          = [for port in local.ports : port.port]
  lb_internal_ip = azurerm_lb.loadbalancer.frontend_ip_configuration[0].private_ip_address
  tags           = var.additional_tags
}

data "azurerm_subscription" "current" {
}

data "azurerm_user_assigned_identity" "uaid" {
  name                = local.uai_name
  resource_group_name = local.uai_resource_group
}
