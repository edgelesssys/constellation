{pkgs ? import <nixpkgs> {}}:
(pkgs.buildFHSUserEnv {
  name = "bazel";
  targetPkgs = pkgs: [
    pkgs.bazel_6
    pkgs.glibc
    pkgs.gcc
    pkgs.jdk11 # TODO(katexochen): investigate why our build chain doesn't work on NixOS
    pkgs.libxcrypt # TODO(katexochen): investigate why our build chain doesn't work on NixOS
  ];
})
.env
