## [E2E upgrade test]((https://github.com/edgelesssys/constellation/actions/workflows/e2e-upgrade.yml)
Make sure to set the correct parameters to avoid late failures:
- it's easiest to use the latest CLI version, because then you can omit all other fields. This works because the devbuild is tagged with the next release version and hence is compatible.
- if using an older CLI version:
  - the last field about simulating a patch-upgrade must have a minor version that is smaller by one compared to the next release
  - the image version must match the patch field version
