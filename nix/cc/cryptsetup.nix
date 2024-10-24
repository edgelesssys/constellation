{ pkgs, pkgsLinux, buildEnv, closureInfo }:
let
  lib = pkgs.lib;
  cc = pkgsLinux.stdenv.cc;
  packages = [ pkgsLinux.cryptsetup.out pkgsLinux.cryptsetup.dev ];
  closure = builtins.toString (lib.strings.splitString "\n" (builtins.readFile "${closureInfo {rootPaths = packages;}}/store-paths"));
  rpath = pkgs.lib.makeLibraryPath [ pkgsLinux.cryptsetup pkgsLinux.glibc pkgsLinux.libgcc.lib ];
in
pkgs.symlinkJoin {
  name = "cryptsetup";
  paths = packages;
  buildInputs = packages;
  postBuild = ''
    tar -cf $out/closure.tar --mtime="@$SOURCE_DATE_EPOCH" --sort=name --hard-dereference ${closure}
    echo "${rpath}" > $out/rpath
    cp ${cc}/nix-support/dynamic-linker $out/dynamic-linker
  '';
}
