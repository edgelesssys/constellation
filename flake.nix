{
  description = "Constellation";

  inputs = {
    nixpkgsUnstable = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs =
    { self
    , nixpkgsUnstable
    , flake-utils
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgsUnstable = import nixpkgsUnstable { inherit system; };

      callPackage = pkgsUnstable.callPackage;

      mkosiDev = (pkgsUnstable.mkosi.overrideAttrs (oldAttrs: rec {
        # TODO(malt3): remove patch once merged and released upstream (systemd/mkosi#2163)
        src = pkgsUnstable.fetchFromGitHub {
          owner = "systemd";
          repo = "mkosi";
          rev = "0a4e1276af191e64c57cd3c1ab50271ab3b706c8";
          hash = "sha256-vcXa+oc+5hpWPQvzloNPTM8ra7ShUTLtBiQjqoAB09Y=";
        };
        propagatedBuildInputs = oldAttrs.propagatedBuildInputs ++ (with pkgsUnstable;  [
          # package management
          dnf5
          rpm

          # filesystem tools
          bash # sh and bash for sandboxed scripts
          coreutils # cp, rm, mkdir, chmod, chown, chgrp, ln, readlink, stat, ...
          btrfs-progs # btrfs subvolumes
          gnutar # tar
          squashfsTools # mksquashfs
          dosfstools # mkfs.vfat
          mtools # mcopy
          cryptsetup # dm-verity
          util-linux # flock
          kmod # depmod
          cpio # cpio
          zstd # zstd
          xz # xz
        ]);
      }));

      openssl-static = pkgsUnstable.openssl.override { static = true; };

    in
    {
      packages.mkosi = mkosiDev;

      packages.openssl = callPackage ./nix/cc/openssl.nix { pkgs = pkgsUnstable; };

      packages.cryptsetup = callPackage ./nix/cc/cryptsetup.nix { pkgs = pkgsUnstable; pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; }; };

      packages.libvirt = callPackage ./nix/cc/libvirt.nix { pkgs = pkgsUnstable; pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; }; };

      packages.libvirtd_base = callPackage ./nix/container/libvirtd_base.nix { pkgs = pkgsUnstable; pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; }; };

      packages.awscli2 = pkgsUnstable.awscli2;

      packages.bazel_6 = pkgsUnstable.bazel_6;

      packages.createrepo_c = pkgsUnstable.createrepo_c;

      packages.dnf5 = pkgsUnstable.dnf5;

      devShells.default = import ./nix/shells/default.nix { pkgs = pkgsUnstable; };

      formatter = nixpkgsUnstable.legacyPackages.${system}.nixpkgs-fmt;
    });
}
