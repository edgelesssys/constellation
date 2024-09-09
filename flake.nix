{
  description = "Constellation";

  inputs = {
    nixpkgsUnstable = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };
    # TODO(msanft): Remove once https://github.com/NixOS/nixpkgs/commit/c429fa2ffa21229eeadbe37c11a47aff35f53ce0
    # lands in nixpkgs-unstable.
    nixpkgsBazel = {
      url = "github:NixOS/nixpkgs/c429fa2ffa21229eeadbe37c11a47aff35f53ce0";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    uplosi = {
      url = "github:edgelesssys/uplosi";
      inputs.nixpkgs.follows = "nixpkgsUnstable";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgsUnstable,
      nixpkgsBazel,
      flake-utils,
      uplosi,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgsUnstable = import nixpkgsUnstable { inherit system; };

        bazelPkgsUnstable = import nixpkgsBazel { inherit system; };

        callPackage = pkgsUnstable.callPackage;

        mkosiDev = (
          pkgsUnstable.mkosi.overrideAttrs (oldAttrs: rec {
            propagatedBuildInputs =
              oldAttrs.propagatedBuildInputs
              ++ (with pkgsUnstable; [
                # package management
                dnf5
                rpm
                createrepo_c

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

                # utils
                gnused # sed
                gnugrep # grep
              ]);
          })
        );

        uplosiDev = uplosi.outputs.packages."${system}".uplosi;

        openssl-static = pkgsUnstable.openssl.override { static = true; };

        bazel_7 = bazelPkgsUnstable.callPackage ./nix/packages/bazel.nix {
          pkgs = bazelPkgsUnstable;
          nixpkgs = nixpkgsBazel;
        };

      in
      {
        packages.mkosi = mkosiDev;

        packages.uplosi = uplosiDev;

        packages.openssl = callPackage ./nix/cc/openssl.nix { pkgs = pkgsUnstable; };

        packages.cryptsetup = callPackage ./nix/cc/cryptsetup.nix {
          pkgs = pkgsUnstable;
          pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; };
        };

        packages.libvirt = callPackage ./nix/cc/libvirt.nix {
          pkgs = pkgsUnstable;
          pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; };
        };

        packages.libvirtd_base = callPackage ./nix/container/libvirtd_base.nix {
          pkgs = pkgsUnstable;
          pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; };
        };

        packages.vpn = callPackage ./nix/container/vpn/vpn.nix {
          pkgs = pkgsUnstable;
          pkgsLinux = import nixpkgsUnstable { system = "x86_64-linux"; };
        };

        packages.awscli2 = pkgsUnstable.awscli2;

        packages.bazel_7 = bazel_7;

        packages.createrepo_c = pkgsUnstable.createrepo_c;

        packages.dnf5 = pkgsUnstable.dnf5;

        devShells.default = callPackage ./nix/shells/default.nix { inherit bazel_7; };

        formatter = nixpkgsUnstable.legacyPackages.${system}.nixpkgs-fmt;
      }
    );
}
