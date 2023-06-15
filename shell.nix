{pkgs ? import <nixpkgs> {}}:
(pkgs.buildFHSUserEnv {
  name = "bazel";
  targetPkgs = pkgs: [
    pkgs.git
    pkgs.go
    pkgs.bazel_6
    pkgs.glibc
    pkgs.gcc
    pkgs.jdk11 # TODO(katexochen): investigate why our build chain doesn't work on NixOS
    pkgs.libxcrypt-legacy # TODO(malt3): python toolchain depends on this. Remove once https://github.com/bazelbuild/rules_python/issues/1211 is resolved
  ];
})
.env
