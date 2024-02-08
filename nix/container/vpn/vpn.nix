{ pkgs
, pkgsLinux
, stdenv
}:
let
  passwd = pkgs.writeTextDir "etc/passwd" ''
    root:x:0:0:root:/root:/bin/sh
    nobody:x:65534:65534:Kernel Overflow User:/:/sbin/nologin
  '';
  group = pkgs.writeTextDir "etc/group" ''
    root:x:0:
    nobody:x:65534:
  '';

  strongswanScript = pkgsLinux.writeShellApplication {
    name = "strongswan.sh";
    runtimeInputs = with pkgsLinux; [
      coreutils
      strongswan
    ];
    text = ./strongswan.sh;
  };

  sidecarScript = pkgsLinux.writeShellApplication {
    name = "sidecar.sh";
    runtimeInputs = with pkgsLinux; [
      coreutils
      iproute2
      jq
      util-linux
      procps
    ];
    text = ./sidecar.sh;
  };

  operatorScript = pkgsLinux.writeShellApplication {
    name = "operator.sh";
    runtimeInputs = with pkgsLinux; [
      coreutils
      kubernetes
      jq
    ];
    text = ./operator.sh;
  };

  image = pkgs.dockerTools.buildImage {
    name = "ghcr.io/edgelesssys/constellation/vpn";
    copyToRoot = with pkgsLinux.dockerTools; [
      passwd
      group
      strongswanScript
      sidecarScript
      operatorScript
      binSh
    ];
    config = {
      Cmd = [ "/bin/entrypoint.sh" ];
    };
  };

in

stdenv.mkDerivation {
  name = "image";

  src = image;

  buildInputs = with pkgs; [ gnutar jq ];

  # unpackPhase = "true";

  installPhase = ''
    mkdir -p "$out/tmp"
    pushd "$out/tmp"
    tar -xf ${image}
    layer="$(jq -r '.[0].Layers[0]' <manifest.json)"
    chmod -R u+w "."
    mv "$layer" "$out/layer.tar"
    popd
    rm -rf -- "$out/tmp"
  '';

}
