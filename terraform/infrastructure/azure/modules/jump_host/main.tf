resource "azurerm_linux_virtual_machine" "jump_host" {
  name                = "${var.base_name}-jump-host"
  resource_group_name = var.resource_group
  location            = var.location
  size                = "Standard_D2as_v5"
  tags                = var.tags

  network_interface_ids = [
    azurerm_network_interface.jump_host.id,
  ]

  admin_username = "adminuser"

  admin_ssh_key {
    username   = "adminuser"
    public_key = tls_private_key.ssh_key.public_key_openssh
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  boot_diagnostics {
    storage_account_uri = null
  }

  user_data = base64encode(<<EOF
#!/bin/bash
set -x

# Uncomment to create user with password
# useradd -m user
# usermod -aG sudo user
# usermod --shell /bin/bash user
# sh -c "echo \"user:pass\" | chpasswd"

sysctl -w net.ipv4.ip_forward=1
sysctl -p

internal_ip=$(ip route get 8.8.8.8 | grep -oP 'src \K[^ ]+')

lb_ip=${var.lb_internal_ip}
if [[ ! $${lb_ip} =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  lb_ip=$(dig +short ${var.lb_internal_ip})
fi

%{for port in var.ports~}
iptables -t nat -A PREROUTING -p tcp --dport ${port} -j DNAT --to-destination $${lb_ip}:${port}
iptables -t nat -A POSTROUTING -p tcp -d $${lb_ip} --dport ${port} -j SNAT --to-source $${internal_ip}
%{endfor~}
EOF
  )
}

resource "azurerm_network_interface" "jump_host" {
  name                = "${var.base_name}-jump-host"
  resource_group_name = var.resource_group
  location            = var.location
  tags                = var.tags

  ip_configuration {
    name                          = "public"
    subnet_id                     = var.subnet_id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.jump_host.id
  }
}

resource "azurerm_public_ip" "jump_host" {
  name                = "${var.base_name}-jump-host"
  resource_group_name = var.resource_group
  location            = var.location
  allocation_method   = "Dynamic"
  tags                = var.tags
}

resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
