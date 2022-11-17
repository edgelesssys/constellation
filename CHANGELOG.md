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

### Changed
<!-- For changes in existing functionality.  -->

### Deprecated
<!-- For soon-to-be removed features. -->

### Removed
<!-- For now removed features. -->
- `access-manager` was removed from code base. K8s native way to SSH into nodes documented.

### Fixed

### Security

## [2.2.2] - 2022-11-17

### Fixed

- `constellation create` on GCP now always uses the local default credentials.
- A release process error encountered in v2.2.1. This led to a broken QEMU-based Constellation deployment, where PCR[8] didn't match.

## [2.2.1] - 2022-11-16

### Changed

- Increase timeout for `constellation config fetch-measurements` from 3 seconds to 60 seconds.
- Consistently log CLI warnings and errors to `stderr`.

### Security

Vulnerabilities in `kube-apiserver` fixed by upgrading to v1.23.14, v1.24.8 and v1.25.4:
- [CVE-2022-3162](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-3162)
- [CVE-2022-3294](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-3294)

## [2.2.0] - 2022-11-08

### Added

- Sign generated SBOMs and store container image SBOMs in registry for easier usage.
- Support for Constellation on AWS.
- Constellation Kubernetes services are now managed using Helm.
- Use tags to mark all applicable resources using a Constellation's UID on Azure.
- Use labels to mark all applicable resources using a Constellation's UID on GCP.

### Changed

- Verify measurements using [Rekor](https://github.com/sigstore/rekor) transparency log.
- The `constellation create` on Azure now uses Terraform to create and destroy cloud resources.
- Constellation OS images are now based on Fedora directly and are built using [mkosi](https://github.com/systemd/mkosi).
- `constellation terminate` will now prompt the user for confirmation before destroying any resources (can be skipped with `--yes`).
- Use the `constellation-role` tag instead of `role` to indicate an instance's role on Azure.
- Use labels instead of metadata to apply the `constellation-uid` and `constellation-role` tags on GCP.

### Deprecated

- `access-manager` is no longer deployed.

### Removed

- `endpoint` flag of `constellation init`. IP is now always taken from the `constellation-id.json` file.
- `constellation-state.json` file won't be created anymore. Resources are now managed through Terraform.

### Fixed

### Security

### Internal

## [2.1.0] - 2022-10-07

### Added

- MiniConstellation: Try out Constellation locally without any cloud subscription required just with one command: `constellation mini up`
- Loadbalancer for control-plane recovery
- K8s conformance mode
- Local cluster creation based on QEMU
- Verification of Azure trusted launch attestation keys
- Kubernetes version v1.25 is now fully supported.
- Enabled Konnectivity.

### Changed
<!-- For changes in existing functionality.  -->
- Autoscaling is now directly managed inside Kubernetes, by the Constellation node operator.
- The `constellation create` on GCP now uses Terraform to create and destroy cloud resources.
- GCP instances are now created without public IPs by default.
- Kubernetes default version used in Constellation is now v1.24.

### Deprecated
<!-- For soon-to-be removed features. -->
### Removed
<!-- For now removed features. -->
- CLI options for autoscaling, as this is now managed inside Kubernetes.
- Kubernetes version v1.22 is no longer supported.

### Fixed

### Security
Vulnerability inside the Go standard library fixed by updating to Go 1.19.2:
- [GO-2022-1037](https://pkg.go.dev/vuln/GO-2022-1037) ([CVE-2022-2879](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-2879))
- [GO-2022-1038](https://pkg.go.dev/vuln/GO-2022-1038) ([CVE-2022-2880](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-2880))
- [GO-2022-0969](https://pkg.go.dev/vuln/GO-2022-0969) ([CVE-2022-27664](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-27664))

### Internal

## [2.0.0] - 2022-09-12

Initial release of Constellation.

[Unreleased]: https://github.com/edgelesssys/constellation/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/edgelesssys/constellation/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/edgelesssys/constellation/releases/tag/v2.0.0
