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

      mkosiDev = (pkgsUnstable.mkosi.overrideAttrs (oldAttrs: rec {
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
        ]);
      }));

      openssl-static = pkgsUnstable.openssl.override { static = true; };

    in
    {
      packages.mkosi = mkosiDev;

      packages.openssl = pkgsUnstable.symlinkJoin {
        name = "openssl";
        paths = [ openssl-static.out openssl-static.dev ];
      };

      packages.awscli2 = pkgsUnstable.awscli2;

      packages.createrepo_c = pkgsUnstable.createrepo_c;

      packages.dnf5 = pkgsUnstable.dnf5;

      devShells.default = import ./nix/shells/default.nix { pkgs = pkgsUnstable; };

      formatter = nixpkgsUnstable.legacyPackages.${system}.nixpkgs-fmt;
    });
}
