terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.33.0"
    }
  }
}

data "google_compute_image" "image_ubuntu" {
  family  = "ubuntu-2204-lts"
  project = "ubuntu-os-cloud"
}

resource "google_compute_instance" "vm_instance" {
  name         = "${var.base_name}-jumphost"
  machine_type = "n2d-standard-4"
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = data.google_compute_image.image_ubuntu.self_link
    }
  }

  network_interface {
    subnetwork = var.subnetwork
    access_config {
    }
  }

  service_account {
    scopes = ["compute-ro"]
  }

  labels = var.labels

  metadata = {
    serial-port-enable = "TRUE"
  }

  metadata_startup_script = <<EOF
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
iptables -t nat -A PREROUTING -p tcp --dport ${port} -j DNAT --to-destination ${var.lb_internal_ip}:${port}
iptables -t nat -A POSTROUTING -p tcp -d ${var.lb_internal_ip} --dport ${port} -j SNAT --to-source $${internal_ip}
%{endfor~}
EOF
}
