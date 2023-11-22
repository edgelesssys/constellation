{ pkgs }:
pkgs.symlinkJoin {
  name = "cryptsetup";
  paths = [ pkgs.cryptsetup.out pkgs.cryptsetup.dev ];
}
