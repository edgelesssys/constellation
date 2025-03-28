{
  description = "Constellation";

  inputs = {
    nixpkgsUnstable = {
      # TODO(msanft): Go back to upstream once the following PR lands:
      # https://github.com/NixOS/nixpkgs/pull/395114
      url = "github:msanft/nixpkgs/msanft/mkosi/fix";
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
        overlay = final: prev: {
          rpm = prev.rpm.overrideAttrs (old: {
            nativeBuildInputs = old.nativeBuildInputs ++ [ prev.makeWrapper ];
            postFixup = ''
              wrapProgram $out/lib/rpm/sysusers.sh \
                --set PATH ${
                  prev.lib.makeBinPath (
                    with prev;
                    [
                      coreutils
                      findutils
                      su.out
                      gnugrep
                    ]
                  )
                }
            '';
          });

          # dnf5 assumes a TTY with a very small width by default, truncating its output instead of line-wrapping
          # it. Force it to use more VT columns to avoid this, and make debugging errors easier.
          dnf5-stub = prev.writeScriptBin "dnf5" ''
            #!/usr/bin/env bash
            FORCE_COLUMNS=200 ${final.dnf5}/bin/dnf5 $@
          '';
        };

        pkgsUnstable = import nixpkgsUnstable {
          inherit system;
          overlays = [ overlay ];
        };

        callPackage = pkgsUnstable.callPackage;

        mkosiDev = (
          pkgsUnstable.mkosi.override {
            extraDeps = (
              with pkgsUnstable;
              [
                # package management
                dnf5-stub
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
              ]
            );
          }
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
