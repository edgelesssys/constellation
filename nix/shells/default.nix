{ pkgs, ... }:

(pkgs.buildFHSUserEnv {
  name = "bazel-userenv";
  targetPkgs = with pkgs; pkgs: [
    bazel_6
    glibc
    git
  ];
  #extraBwrapArgs = [
  #  "--bind-try ~/.cache ~/.cache"
  #];
}).env
