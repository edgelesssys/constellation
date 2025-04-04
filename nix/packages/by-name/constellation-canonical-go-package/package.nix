{
  lib,
  buildGo124Module,
}:
buildGo124Module {
  pname = "constellation-canonical-go-package";
  version = lib.constellationVersion;

  src = lib.constellationRepoRootSrc [
    "go.mod"
    "go.sum"
  ];

  vendorHash = "sha256-McWiTTz1HTdG3x0LI87CF6oTRFtxSiV3LCCBJb9YG4U=";

  doCheck = false;

  proxyVendor = true;
}
