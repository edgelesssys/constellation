# Changelog
All notable changes to Constellation will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Early boot logging for Cloud Provider: GCP & Azure
- Added `constellation-access-manager`, allowing users to manage SSH users over a ConfigMap. This allows persistent & dynamic management of SSH users on multiple nodes, even after a reboot.

### Changed

### Removed

### Fixed

### Security
- GCP WireGuard encryption via cilium

### Internal
- Added `constellation-activation-service`, offloading new Kubernetes node activation from monolithic Coordinator to Kubernetes native micro-service

## [1.2.0] - 2022-06-02
### Added

### Changed
replaced flannel CNI with cilium

### Removed

### Fixed

### Security

### Internal

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

[Unreleased]: https://github.com/edgelesssys/constellation/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/edgelesssys/constellation/releases/tag/v1.2.0
[1.1.0]: https://github.com/edgelesssys/constellation/releases/tag/v1.1.0
[1.0.0]: https://github.com/edgelesssys/constellation/releases/tag/v1.0.0
