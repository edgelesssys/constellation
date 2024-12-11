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
    {
      self,
      nixpkgsUnstable,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgsUnstable = import nixpkgsUnstable { inherit system; };

        callPackage = pkgsUnstable.callPackage;

        mkosiDev = (
          pkgsUnstable.mkosi.overrideAttrs (oldAttrs: {
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
      in
      {
        packages.mkosi = mkosiDev;

        packages.uplosi = pkgsUnstable.uplosi;

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

        packages.createrepo_c = pkgsUnstable.createrepo_c;

        packages.dnf5 = pkgsUnstable.dnf5;

        devShells.default = callPackage ./nix/shells/default.nix { };

        formatter = nixpkgsUnstable.legacyPackages.${system}.nixpkgs-fmt;
      }
    );
}
