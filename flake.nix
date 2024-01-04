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
          rev = "abf22cdc6ccb13f2cd84679ede77231455ec6813";
          hash = "sha256-njtYWSXSLMcn6AtGfAeL/ncZQ6g+Vgpe7EaKLkzAOl4=";
        };
        propagatedBuildInputs = oldAttrs.propagatedBuildInputs ++ (with pkgsUnstable;  [
          # package management
          dnf5
          rpm

          # filesystem tools
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

      uplosiDev = (pkgsUnstable.uplosi.overrideAttrs (oldAttrs: rec {
        src = pkgsUnstable.fetchFromGitHub {
          owner = "edgelesssys";
          repo = "uplosi";
          rev = "0190e8c548b5811066b7e2d9db5e3167f51c005f";
          hash = "sha256-AHj3XTX+vd8QP4hWGPAt2iJnrIGoiH61UgQMK7vlYU0=";
        };
      }));

      openssl-static = pkgsUnstable.openssl.override { static = true; };

    in
    {
      packages.mkosi = mkosiDev;

      packages.uplosi = uplosiDev;

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
