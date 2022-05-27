# Changelog
All notable changes to Constellation will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- CLI
  - Command `constellation recover` to re-initialize a completely stopped cluster.
  - Command `constellation config generate` to generate a default configuration file for a specific cloud provider.
- CSI
  - Option to enable dm-integrity in a StorageClass.
  - Support volume expansion.
  - Support volume snapshots.
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

[Unreleased]: https://github.com/edgelesssys/constellation/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/edgelesssys/constellation/releases/tag/v1.0.0
