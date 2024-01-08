{ pkgs, ... }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    bazel_7
    git
  ];
}
