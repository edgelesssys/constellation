# Versions and support policy

This page details which guarantees Constellation makes regarding versions, compatibility, and life cycle.

All released components of Constellation use a three-digit version number, of the form `v<MAJOR>.<MINOR>.<PATCH>`.

## Release cadence

All [components](components.md) of Constellation are released in lock step on the first Tuesday of every month. This release primarily introduces new features, but may also include security or performance improvements. The `MINOR` version will be incremented as part of this release.

Additional `PATCH` releases may be created on demand, to fix security issues or bugs before the next `MINOR` release window.

New releases are published on [GitHub](https://github.com/edgelesssys/constellation/releases).

### Kubernetes support policy

Constellation is aligned to the [version support policy of Kubernetes](https://kubernetes.io/releases/version-skew-policy/#supported-versions), and therefore supports the most recent three minor versions.
