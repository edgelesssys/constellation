let
  pkgs = import <nixpkgs> { };
  goAttr = pkgs.go_1_21.overrideAttrs (_: rec {
    version = "1.21.5";
    src = pkgs.fetchurl {
      url = "https://go.dev/dl/go${version}.src.tar.gz";
      hash = "sha256-KFy730tubmLtWPNw8/bYwwgl1uVsWFPGbTwjvNsJ2xk=";
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
