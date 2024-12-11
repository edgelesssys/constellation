terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "3.0.0"
    }
  }
}

resource "openstack_lb_listener_v2" "listener" {
  name            = var.name
  protocol        = "TCP"
  protocol_port   = var.port
  loadbalancer_id = var.loadbalancer_id
}

resource "openstack_lb_pool_v2" "pool" {
  name        = var.name
  protocol    = "TCP"
  lb_method   = "ROUND_ROBIN"
  listener_id = openstack_lb_listener_v2.listener.id
}

resource "openstack_lb_member_v2" "member" {
  count         = length(var.member_ips)
  name          = format("${var.name}-member-%02d", count.index + 1)
  address       = var.member_ips[count.index]
  protocol_port = var.port
  pool_id       = openstack_lb_pool_v2.pool.id
  subnet_id     = var.subnet_id
}

resource "openstack_lb_monitor_v2" "k8s_api" {
  name        = var.name
  pool_id     = openstack_lb_pool_v2.pool.id
  type        = "TCP"
  delay       = 2
  timeout     = 2
  max_retries = 2
}
