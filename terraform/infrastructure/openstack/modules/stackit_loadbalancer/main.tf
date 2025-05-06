terraform {
  required_providers {
    stackit = {
      source  = "stackitcloud/stackit"
      version = "0.51.0"
    }
  }
}

resource "stackit_loadbalancer" "loadbalancer" {
  project_id = var.stackit_project_id
  name       = "${var.name}-lb"
  target_pools = [
    for portName, port in var.ports : {
      name        = "target-pool-${portName}"
      target_port = port
      targets = [
        for ip in var.member_ips : {
          display_name = "target-${portName}"
          ip           = ip
        }
      ]
      active_health_check = {
        healthy_threshold   = 10
        interval            = "3s"
        interval_jitter     = "3s"
        timeout             = "3s"
        unhealthy_threshold = 10
      }
    }
  ]
  listeners = [
    for portName, port in var.ports : {
      name        = "listener-${portName}"
      port        = port
      protocol    = "PROTOCOL_TCP"
      target_pool = "target-pool-${portName}"
    }
  ]
  networks = [
    {
      network_id = var.network_id
      role       = "ROLE_LISTENERS_AND_TARGETS"
    }
  ]
  external_address = var.external_address
}
