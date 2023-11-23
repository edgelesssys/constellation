{ pkgs, buildEnv, closureInfo }:
let
  lib = pkgs.lib;
  packages = [ pkgs.libvirt ];
  closure = builtins.toString (lib.strings.splitString "\n" (builtins.readFile "${closureInfo {rootPaths = packages;}}/store-paths"));
  rpath = pkgs.lib.makeLibraryPath [ pkgs.libvirt pkgs.glib pkgs.libxml2 pkgs.readline pkgs.glibc pkgs.libgcc.lib ];
in
pkgs.symlinkJoin {
  name = "libvirt";
  paths = [ pkgs.libvirt ];
  buildInputs = packages;
  postBuild = ''
    tar -cf $out/closure.tar ${closure}
    tar --transform 's+^./+bin/+' -cf $out/bin-linktree.tar -C $out/bin .
    echo "${rpath}" > $out/rpath
  '';
}
