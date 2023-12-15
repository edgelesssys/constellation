data "constellation_attestation" "foo" {} # Fill accordingly for the CSP and attestation variant

data "constellation_image" "bar" {} # Fill accordingly for the CSP

resource "random_bytes" "master_secret" {
  length = 32
}

resource "random_bytes" "master_secret_salt" {
  length = 32
}

resource "random_bytes" "measurement_salt" {
  length = 32
}

resource "constellation_cluster" "azure_example" {
  csp                                = "azure"
  constellation_microservice_version = "vX.Y.Z"
  name                               = "constell"
  uid                                = "..."
  image                              = data.constellation_image.bar.image
  attestation                        = data.constellation_attestation.foo.attestation
  init_secret                        = "..."
  master_secret                      = random_bytes.master_secret.hex
  master_secret_salt                 = random_bytes.master_secret_salt.hex
  measurement_salt                   = random_bytes.measurement_salt.hex
  out_of_cluster_endpoint            = "123.123.123.123"
  azure = {
    tenant_id                   = "..."
    subscription_id             = "..."
    uami_client_id              = "..."
    uami_resource_id            = "..."
    location                    = "..."
    resource_group              = "..."
    load_balancer_name          = "..."
    network_security_group_name = "..."
  }
  network_config = {
    ip_cidr_node    = "192.168.176.0/20"
    ip_cidr_service = "10.96.0.0/12"
  }
}
