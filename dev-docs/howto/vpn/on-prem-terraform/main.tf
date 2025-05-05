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
  features {}
  subscription_id = var.subscription_id
  # This enables all resource providers.
  # In the future, we might want to use `resource_providers_to_register` to registers just the ones we need.
  resource_provider_registrations = "all"
}

locals {
  username = "azureadmin"
}

resource "random_pet" "rg_name" {
  prefix = var.name_prefix
}

resource "azurerm_resource_group" "rg" {
  location = var.resource_group_location
  name     = random_pet.rg_name.id
}

# Create virtual network
resource "azurerm_virtual_network" "network" {
  name                = "network"
  address_space       = [var.local_ts]
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

# Create subnet
resource "azurerm_subnet" "subnet" {
  name                 = "subnet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = [cidrsubnet(var.local_ts, 8, 0)]

}

resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Create public IPs
resource "azurerm_public_ip" "pubIP" {
  name                = "publicIP"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Dynamic"
}

# Create Network Security Group and rule
resource "azurerm_network_security_group" "security_group" {
  name                = "secuityGroup"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "strongSwan_500"
    priority                   = 1002
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Udp"
    source_port_range          = "*"
    destination_port_range     = "500"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
  security_rule {
    name                       = "strongSwan_4500"
    priority                   = 1003
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Udp"
    source_port_range          = "*"
    destination_port_range     = "4500"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_route_table" "route_table" {
  name                          = "vpn-routes"
  location                      = azurerm_resource_group.rg.location
  resource_group_name           = azurerm_resource_group.rg.name
  bgp_route_propagation_enabled = false

  dynamic "route" {
    for_each = var.remote_ts
    content {
      name                   = "route-${route.key}"
      address_prefix         = route.value
      next_hop_type          = "VirtualAppliance"
      next_hop_in_ip_address = azurerm_network_interface.public_nic.private_ip_address
    }
  }
}

resource "azurerm_subnet_route_table_association" "route_table_association" {
  subnet_id      = azurerm_subnet.subnet.id
  route_table_id = azurerm_route_table.route_table.id
}


# Create network interface
resource "azurerm_network_interface" "public_nic" {
  name                = "public-nic"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "my_nic_configuration"
    subnet_id                     = azurerm_subnet.subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.pubIP.id
  }
}

# Connect the security group to the network interface
resource "azurerm_network_interface_security_group_association" "example" {
  network_interface_id      = azurerm_network_interface.public_nic.id
  network_security_group_id = azurerm_network_security_group.security_group.id
}

# Create virtual machine
resource "azurerm_linux_virtual_machine" "public_vm" {
  name                  = "public_vm"
  location              = azurerm_resource_group.rg.location
  resource_group_name   = azurerm_resource_group.rg.name
  network_interface_ids = [azurerm_network_interface.public_nic.id]
  size                  = "Standard_B2ats_v2"

  os_disk {
    name                 = "disk_public_vm"
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  computer_name  = "hostname"
  admin_username = local.username

  admin_ssh_key {
    username   = local.username
    public_key = tls_private_key.ssh_key.public_key_openssh
  }

  boot_diagnostics {
  }

  user_data = base64encode(<<EOF
#!/bin/bash
set -x

apt-get update
apt-get install strongswan-charon strongswan-swanctl -y


cat <<'EOT' >> /etc/strongswan.d/charon-logging.conf
charon {
  filelog {
    stderr {
      time_format = %b %e %T
      ike_name = yes
      default = 1
      ike = 2
      flush_line = yes
    }
  }
}
EOT


cat <<'EOT' >> /etc/swanctl/conf.d/constellation.conf
connections {
  gw-gw {
    remote_addrs = ${var.remote_addr}

    local {
        auth = psk
    }
    remote {
        auth = psk
    }
    children {
        net-net {
          local_ts  = ${var.local_ts}
          remote_ts = ${join(",", var.remote_ts)}

          start_action = trap
        }
    }
  }
}

secrets {
  ike {
    secret = ${var.ike_psk}
  }
}
EOT

cat <<'EOT' >> /home/${local.username}/restart-and-reload-strongswan.sh
#!/bin/sh

# Restart charon daemon
ipsec restart

sleep 5

# Load all the config files
swanctl --load-all

echo "You now should be able to ping and curl the remote network (Pod IPs and Services)"

EOT

chmod +x /home/${local.username}/restart-and-reload-strongswan.sh
sysctl -w net.ipv4.ip_forward=1

EOF
  )
}

resource "azurerm_network_interface" "private_nic" {
  name                = "private-nic"
  location            = var.resource_group_location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.subnet.id
    private_ip_address_allocation = "Dynamic"
  }
}

# Create virtual machine
resource "azurerm_linux_virtual_machine" "private_vm" {
  name                  = "private_vm"
  location              = azurerm_resource_group.rg.location
  resource_group_name   = azurerm_resource_group.rg.name
  network_interface_ids = [azurerm_network_interface.private_nic.id]
  size                  = "Standard_B2ats_v2"

  os_disk {
    name                 = "disk_private_vm"
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  computer_name  = "hostname"
  admin_username = local.username

  admin_ssh_key {
    username   = local.username
    public_key = tls_private_key.ssh_key.public_key_openssh
  }

  boot_diagnostics {
  }
}
