# Constellation Terraform Provider

The Constellation Terraform Provider allows its user to manage the full lifecycle of a Constellation cluster -- namely creation, updates, and deletion -- via Terraform.

## Design Goals

- The User needs to be able to perform Creation, Upgrades and Deletion of a Cluster through the provider.
- Infrastructure provisioning should explicitly **not** be performed by the Constellation Terraform provider.
- The User needs to receive some sort of Access Token to the Cluster (e.g. a Kubeconfig, whether long- or short-lived)
when applying a configuration containing the provider and the resources listed below.

## Terraform Configuration Layout for the Constellation Resource

This resembles an examplary configuration of a Constellation cluster through Terraform. While the exact naming and struct layout
of the individual components may change until the implementation, the structure of the overall configuration looks like this:

```hcl
terraform {
  required_providers {
    constellation = {
      source  = "tbd/constellation"
      version = "2.13.0"
    }
  }
}

provider "constellation" { }

resource "constellation_cluster" "foo" {
    uid = "bar"
    name = "baz"
    image = data.constellation_image.ref # or provide manually crafted values
    kubernetes_version = "v1.27.6"
    constellation_microservice_version = "lockstep" # or pinned to a specific version
    debug = false
    init_endpoint = "10.10.10.10" # should use public ip of LB resource, ideally also provisioned through TF
    kubernetes_api_endpoint = "10.10.10.10" # should use public ip of LB resource, ideally also provisioned through TF
    extra_microservices = {
        csi_driver = true
        # + more
        # possiblly also constellation microservices with version and maybe service options,
        # which would make constellation_microservice_version obsolete.
        # exact API TBD
    }
    master_secret = "foo" # updating this would force recreation of the cluster
    init_secret = "bar" # maybe derive from master_secret, updating this would force recreation of the cluster
    network_config = {
        # TBD
        # should contain CIDRs for pod network, service cidr, node network... for Cilium
        # and in-cluster Kubernetes API endpoint, e.g. for Kubelets
    }
    attestation = data.constellation_attestation.attestation # or provide manually crafted values
}

# constellation_cluster provides:
# constellation_cluster.foo.kubeconfig
# constellation_cluster.foo.owner_id
# constellation_cluster.foo.cluster_id


data "constellation_attestation" "foo" {
    attestation_variant = "GCP-SEV-ES"
    image_version = "v2.13.0" # or "lockstep"
    maa_url = "https://www.example.com" # optional, for Azure only
}

# constellation_attestation provides:
# data.constellation_attestation.foo.measurements
# data.constellation_attestation.foo.attestation


data "constellation_image" "foo" {
    image_version = "v2.21.0" # or "lockstep"
    attestation_variant = "GCP-SEV-ES"
    csp = "GCP"
    region = "us-central1" # optional, for AWS only
}

# constellation_image provides:
# constellation_image.foo.reference (CSP-specific image reference)
```

The `constellation_cluster` resource is the main resource implemented by the provider.
It resembles a Constellation cluster with a specific configuration.
Applying it will create the cluster if not existing *or* if the required updates can't be performed in-place (e.g. changing the master secret),
updates it with the according configuration if already existing, or deletes it if not present in the configuration but in the state.

The "constellation_attestation" and "constellation_image" objects are [data sources](https://developer.hashicorp.com/terraform/language/data-sources),
which are objects that should be evaluated by the Provider each time the state is refreshed (i.e. each time any Terraform command that evaluates configuration against state),
but have no observable side effects. For image and attestation, this is required as the provider need to evaluate `latest` values or map CSP-agnostic image references (e.g. `v2.13.0`)
to CSP-specific image references (e.g. `/CommunityGalleries/.../Image` for Azure). This is implemented as an idempotent API call and thus has no observable side-effects, but needs
to be re-evaluated as the values returned by the API might change between evaluations.
