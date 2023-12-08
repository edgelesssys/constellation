// Not up-to-date

data "constellation_attestation" "foo" {} # Fill accordingly for the CSP and attestation variant

data "constellation_image" "bar" {} # Fill accordingly for the CSP

resource "constellation_cluster" "aws_example" {
  csp                                = "aws"
  name                               = "constell"
  uid                                = "deadbeef"
  constellation_microservice_version = "vx.y.z"
  image                              = data.constellation_image.bar.image
  attestation                        = data.constellation_attestation.foo.attestation
  init_secret                        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
  master_secret                      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
  master_secret_salt                 = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
  out_of_cluster_endpoint            = "123.123.123.123"
  network_config = {
    ip_cidr_node    = "192.168.176.0/20"
    ip_cidr_service = "10.96.0.0/12"
  }
}
