<!--

Styleguide for this document:

- Sentences should end with a period.
  - This is the keepachangelog style, whereas the Microsoft Style Guide we use for other docs omits periods for short list items.
- Omit the verb if possible.
  - "Early boot logging ..." instead of "Add early boot logging ...".
  - If you need a verb, it should usually be imperative mood (Add instead of Added).
- Items should start with a capital letter.

-->

# Changelog

All notable changes to Constellation will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Support Linux arm64 and macOS (arm64 and amd64) for Constellation CLI.
- Create multiple load balancers to enable load balacing TCP traffic for different backend services. All load balancers currently share the same public IP address.
- Improve rollback on GCP resource termination. You can now terminate multiple times.
- Implement SSH peer to peer distribution between debugd nodes.
- GCP service account can now be managed manually.
- Azure resource group can now be managed manually and can be resused after termination.
- Azure Active Directory client credentials can now be managed manually.
- Resources on Azure are now tagged with the UID of the constellation.
- CoreOS images are publicly available for Azure.
- GCP: Support for higher end N2D standard (128 & 224 vCPUs), *high-mem* and *high-cpu* VMs
- Add `constellation upgrade` to update node images in Constellation.
- Add cilium v1.12.1 with strict mode v2
- Konnectivity is now deployed for secure API server to node/pod/service communication.

### Changed
<!-- For changes in existing functionality.  -->
- Use IP from Constellation ID file in init and verify instead of IPs from state file.
- Terminate now deletes all resources found within the given resource group.
- Change cdbg to use load balancer for deploy.
- cdbg now uses the Constellation config directly and does not require any extra config
- Azure CVMs are attested using SNP attestation
- Replaced kube-proxy with cilium
- VM instance types are now defined in the config, not via a CLI argument
- Config has a `debugCluster` flag required to enable debugd ingress firewall rules and create the required load balancer.

### Deprecated
<!-- For soon-to-be removed features. -->
### Removed
<!-- For now removed features. -->
- Azure Trusted Launch instance types with 2 CPUs (SMT disabled due to Retbleed (CVE-2022-29900)).
- cdbg: Custom systemd service deployment
- No user configurable `ingressFirewall` and `egressFirewall` in the config

### Fixed

### Security
<!-- In case of vulnerabilities. -->
### Internal

## [1.5.0] - 2022-08-19

### Added

- Kubernetes operator for Constellation nodes with ability to update node images.
- CoreOS images are publicly available for GCP.
- Cilium strict pod2pod encryption.
- Add a configurable list of enforced measurements to the config. If an expected measurement can not be verified during attestation, but it is not in the list of enforced measurements, only a warning is logged.
- License check during init

### Changed

- Use Azure CVMs instead of Trusted Launch VMs.
- Parallel resource creation on Azure.

### Fixed

- Fix timeout issue during cilium installation.

### Internal

- Run e2e tests on all supported versions.
- Run e2e tests on latest debug images, instead of release image.
- Upgrade Azure SDK

## [1.4.0] - 2022-08-02

### Added

- Publish measurements for each released coreos-image.
- `constellation config fetch-measurements` to download and verify measurements, and writing them into the config file.
- Configurable Kubernetes version through an entry in `constellation-config.yaml`.
- Kubernetes version 1.24 support.
- Kubernetes version 1.22 support.
- Log disk UUID to cloud logging for recovery.
- Configurable disk type for Azure and GCP.
- Create Kubernetes CA signed kubelet certificates on activation.
- Salt key derivation.
- Integrity protection of state disks.

### Changed

- Nodes add themselves to the cluster after `constellation init` is done. Previously, nodes were asked to join the cluster by the bootstrapper.
- Owner ID and Unique ID are merged into a single value: Cluster ID.
- Streamline logging to only use one logging library, instead of multiple.
- Replace dependency on github.com/willdonnelly/passwd with own implementation.
- Refactor disk-mapper to allow a more streamlined node recovery

### Removed

- User facing WireGuard VPN.

### Fixed

- Correctly wait for `bootstrapper` to come online during `constellation init`.

## [1.3.1] - 2022-07-11

### Changed

- Update default CoreOS image to latest version (1657199013).

### Fixed

- Add load balancer path to Azure deployment so that PCR values can be read.
- Show correct version number in `constellation version`.

### Removed

- Support for Azure `Standard_*_v3` types.

## [1.3.0] - 2022-07-05

### Added

- Early boot logging for GCP and Azure. [[Docs]](https://docs.edgeless.systems/constellation/workflows/troubleshooting#cloud-logging)
- `constellation-access-manager` allows users to manage SSH users over a ConfigMap. Enables persistent and dynamic management of SSH users on multiple nodes, even after a reboot. [[Docs]](https://docs.edgeless.systems/constellation/workflows/ssh)
- GCP-native Kubernetes load balancing. [[Docs]](https://docs.edgeless.systems/constellation/architecture/networking)
- `constellation version` prints more information to aid in troubleshooting. [[Docs]](https://docs.edgeless.systems/constellation/reference/cli#constellation-version)
- Standard logging for all services and CLI, allows users to control output in a consistent manner.
- `constellation-id.json` in Constellation workspace now holds cluster IDs, to reduce required arguments in Constellation commands, e.g., `constellation verify`.

### Changed

- New `constellation-activation-service` offloads Kubernetes node activation from monolithic Coordinator to Kubernetes native micro-service. [[ReadMe]](https://github.com/edgelesssys/constellation/blob/main/activation/README.md)
- Improve user-friendliness of error messages in Constellation CLI.
- Move verification from extracting attestation statements out of aTLS handshake to a dedicated `verify-service` in Kubernetes with gRPC and HTTP endpoints.
- `constellation create` and `constellation verify` do not require specifying the provider anymore (automatically parsed from config)

### Security

- GCP WireGuard encryption via cilium.

### Internal

- Refactore folder structure of repository to better reflect `internal` implementation and public API.
- Extend `goleak` checks to all tests.

## [1.2.0] - 2022-06-02

### Changed

- Replace flannel CNI with Cilium.

## [1.1.0] - 2022-06-02

### Added

- CLI
  - Command `constellation recover` to re-initialize a completely stopped cluster.
  - Command `constellation config generate` to generate a default configuration file for a specific cloud provider.
- CSI
  - Option to enable dm-integrity in a StorageClass.
  - Support volume expansion.
  - Support volume snapshots.
- KMS
  - Deploy Key Management Service (KMS) in Constellation clusters to handle key derivation.
- Option to add SSH users on init.

### Changed

- CLI UX
  - `constellation create` now requires a configuration file. The usual workflow is to run `constellation config generate` first.
  - Consistent command format with at most one argument and named flags otherwise.
  - Display usage when invalid arguments are passed.
  - Add list of instance types to command help.
  - Wording tweaks.
- CLI config
  - Rename dev-config to config.
  - Change format to YAML.
  - Make it self-documenting.
  - Validation.
  - Rename *PCRs* to *Measurements*.

### Removed

- Support for non-CVMs on GCP.

### Fixed

- Pin Kubernetes version deployed by `kubeadm init`.

### Security

- Replace single, never expiring Kubernetes join token with expiring unique tokens.
- Apply CIS benchmark for kubeadm clusterconf and kubelet conf.
- Enable Kubernetes audit log.

### Internal

- Create GCP images in `constellation-images` project so that they can be shared with customers.
- Add customer onboarding docs.
- Add E2E test as Github Action.
- Improvements to local QEMU testing.
- Preparations for mutual ATLS.

## [1.0.0] - 2022-04-28

Initial release of Constellation. With underlying WireGuard and Kubernetes compliant.

[Unreleased]: https://github.com/edgelesssys/constellation/compare/v1.5.0...HEAD
[1.5.0]: https://github.com/edgelesssys/constellation/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/edgelesssys/constellation/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/edgelesssys/constellation/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/edgelesssys/constellation/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/edgelesssys/constellation/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/edgelesssys/constellation/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/edgelesssys/constellation/releases/tag/v1.0.0
