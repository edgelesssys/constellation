let
  pkgs = import <nixpkgs> { };
  goAttr = pkgs.go_1_21.overrideAttrs (_: rec {
    version = "1.21.3";
    src = pkgs.fetchurl {
      url = "https://go.dev/dl/go${version}.src.tar.gz";
      hash = "sha256-GG8rb4yLcE5paCGwmrIEGlwe4T3LwxVqE63PdZMe5Ig=";
    };
  });
in
pkgs.buildEnv
  {
    name = "bazel-go-toolchain";
    paths = [ goAttr ];
    postBuild = ''
      touch $out/ROOT
      ln -s $out/share/go/{api,doc,lib,misc,pkg,src,go.env} $out/
    '';
  } // {
  version = goAttr.version;
}
