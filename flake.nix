{
  description = "Constellation";

  inputs = {
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
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

        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;

          overlays = [
            (_final: prev: (import ./nix/packages { inherit (prev) lib callPackage; }))
            (_final: prev: { lib = prev.lib // (import ./nix/lib { inherit (prev) lib callPackage; }); })
            overlay
          ];
        };

        callPackage = pkgs.callPackage;

        mkosiDev = (
          pkgs.mkosi.override {
            extraDeps = (
              with pkgs;
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
        # Use `legacyPackages` instead of `packages` for the reason explained here:
        # https://github.com/NixOS/nixpkgs/blob/34def00657d7c45c51b0762eb5f5309689a909a5/flake.nix#L138-L156
        # Note that it's *not* a legacy attribute.
        legacyPackages = {
          generate = pkgs.callPackage ./nix/generate.nix { };
        } // pkgs;

        packages.mkosi = mkosiDev;

        packages.uplosi = pkgs.uplosi;

        packages.openssl = callPackage ./nix/cc/openssl.nix { pkgs = pkgs; };

        packages.cryptsetup = callPackage ./nix/cc/cryptsetup.nix {
          pkgs = pkgs;
          pkgsLinux = import nixpkgs { system = "x86_64-linux"; };
        };

        packages.libvirt = callPackage ./nix/cc/libvirt.nix {
          pkgs = pkgs;
          pkgsLinux = import nixpkgs { system = "x86_64-linux"; };
        };

        packages.libvirtd_base = callPackage ./nix/container/libvirtd_base.nix {
          pkgs = pkgs;
          pkgsLinux = import nixpkgs { system = "x86_64-linux"; };
        };

        packages.vpn = callPackage ./nix/container/vpn/vpn.nix {
          pkgs = pkgs;
          pkgsLinux = import nixpkgs { system = "x86_64-linux"; };
        };

        packages.awscli2 = pkgs.awscli2;

        packages.createrepo_c = pkgs.createrepo_c;

        packages.dnf5 = pkgs.dnf5;

        devShells.default = callPackage ./nix/shells/default.nix { };

        formatter = nixpkgs.legacyPackages.${system}.nixpkgs-fmt;
      }
    );
}
