{ pkgs, ... }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    bazel_6
    git
  ];
}
