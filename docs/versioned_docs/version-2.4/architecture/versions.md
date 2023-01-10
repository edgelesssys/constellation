# Versions and support policy

All [components](components.md) of Constellation use a three-digit version number of the form `v<MAJOR>.<MINOR>.<PATCH>`.
The components are released in lock step, usually on the first Tuesday of every month. This release primarily introduces new features, but may also include security or performance improvements. The `MINOR` version will be incremented as part of this release.

Additional `PATCH` releases may be created on demand, to fix security issues or bugs before the next `MINOR` release window.

New releases are published on [GitHub](https://github.com/edgelesssys/constellation/releases).

### Kubernetes support policy

Constellation is aligned to the [version support policy of Kubernetes](https://kubernetes.io/releases/version-skew-policy/#supported-versions), and therefore supports the most recent three minor versions.
