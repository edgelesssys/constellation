# cilium
- Collect values using `helm get values`, merge with values from chart and apply. Needs to be done to pick up any newly introduced values.
- `--reuse-values` is turned off.

# cert-manager
- Collect values using `helm get values`, merge with values from chart and apply. Needs to be done to pick up any newly introduced values.
- installCRDs flag during install and upgrade. This flag is managed by cert-manager.
- WARNING: Print user warning upon upgrading cert-manager that this might break other installations of cert-manager in the cluster.
- `--reuse-values` is turned off.

# operators
- install: crds folder is used to install CRDs
- upgrade: manually update CRDs before doing `helm upgrade`
- `--reuse-values` is turned on.

# constellation-services
- `--reuse-values` is turned on.
