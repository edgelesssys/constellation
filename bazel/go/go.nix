let
  pkgs = import <nixpkgs> { };
  goAttr = pkgs.go_1_21.overrideAttrs (_: rec {
    version = "1.21.4";
    src = pkgs.fetchurl {
      url = "https://go.dev/dl/go${version}.src.tar.gz";
      hash = "sha256-R7Jqg9K2WjwcG8rOJztpvuSaentRaKdgTe09JqN714c=";
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
