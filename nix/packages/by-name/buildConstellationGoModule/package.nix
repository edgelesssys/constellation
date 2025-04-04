# A 'wrapped' Go builder for Constellation, which doesn't require a `vendorHash` to be set in each package.
# Instead, one central vendor hash is set here, and all packages inherit it.

{
  buildGo124Module,
  constellation-canonical-go-package,
}:
args:
(buildGo124Module (
  {
    # We run tests in CI, so don't run them at build time.
    doCheck = false;

    # Disable CGO by default.
    env.CGO_ENABLED = "0";
  }
  // args
)).overrideAttrs
  (_oldAttrs: {
    inherit (constellation-canonical-go-package)
      goModules
      vendorHash
      proxyVendor
      deleteVendor
      ;
  })
