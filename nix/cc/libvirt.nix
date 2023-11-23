{ pkgs, pkgsLinux, buildEnv, closureInfo }:
let
  lib = pkgs.lib;
  cc = pkgsLinux.stdenv.cc;
  packages = [ pkgsLinux.libvirt ];
  closure = builtins.toString (lib.strings.splitString "\n" (builtins.readFile "${closureInfo {rootPaths = packages;}}/store-paths"));
  rpath = pkgs.lib.makeLibraryPath [ pkgsLinux.libvirt pkgsLinux.glib pkgsLinux.libxml2 pkgsLinux.readline pkgsLinux.glibc pkgsLinux.libgcc.lib ];
in
pkgs.symlinkJoin {
  name = "libvirt";
  paths = packages;
  buildInputs = packages;
  postBuild = ''
    tar -cf $out/closure.tar --mtime="@$SOURCE_DATE_EPOCH" --sort=name ${closure}
    tar --transform 's+^./+bin/+' -cf $out/bin-linktree.tar --mtime="@$SOURCE_DATE_EPOCH" --sort=name -C $out/bin .
    echo "${rpath}" > $out/rpath
    cp ${cc}/nix-support/dynamic-linker $out/dynamic-linker
  '';
}
