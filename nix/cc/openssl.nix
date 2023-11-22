{ pkgs }:
let
  openssl-static = pkgs.openssl.override { static = true; };
in
pkgs.symlinkJoin {
  name = "openssl";
  paths = [ openssl-static.out openssl-static.dev ];
}
