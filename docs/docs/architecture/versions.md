# Versions and support policy

All components of Constellation use a three-digit version number of the form `v<MAJOR>.<MINOR>.<PATCH>`.
The components are released in lock step, usually on the first Tuesday of every month. This release primarily introduces new features, but may also include security or performance improvements. The `MINOR` version will be incremented as part of this release.

Additional `PATCH` releases may be created on demand, to fix security issues or bugs before the next `MINOR` release window.

New releases are published on [GitHub](https://github.com/edgelesssys/constellation/releases).

## Kubernetes support policy

Constellation is aligned to the [version support policy of Kubernetes](https://kubernetes.io/releases/version-skew-policy/#supported-versions), and therefore usually supports the most recent three minor versions.
When a new minor version of Kubernetes is released, support is added to the next Constellation release, and that version then supports four Kubernetes versions.
Subsequent Constellation releases drop support for the oldest (and deprecated) Kubernetes version.

The following Kubernetes versions are currently supported:
<!--AUTO_GENERATED_BY_BAZEL-->
<!--DO_NOT_EDIT-->
* v1.28.13
* v1.29.8
* v1.30.4
