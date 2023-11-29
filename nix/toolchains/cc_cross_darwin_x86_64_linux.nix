# cross compiler toolchain for cc rules in Bazel
# execution platform: aarch64-darwin
# target platform: x86_64-linux
# inspired by https://github.com/tweag/rules_nixpkgs/blob/21c4ea481021cb51a6e5d0969b2cee03dba5a637/examples/toolchains/cc_cross_osx_to_linux_amd64/toolchains/osxcross_cc.nix
let
  targetSystem = "x86_64-linux";
  og = import <nixpkgs> { };
  nixpkgs = import <nixpkgs> {
    buildSystem = builtins.currentSystem;
    hostSystem = targetSystem;
    crossSystem = {
      config = targetSystem;
    };
    crossOverlays = [
      (self: super: {
        llvmPackages_11 = super.llvmPackages_11.extend (final: prev: rec {
          libllvm = prev.libllvm.overrideAttrs (old: {
            LDFLAGS = "-L ${super.llvmPackages_11.libcxxabi}/lib";
            nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ og.darwin.cctools ];
          });
          libclang = prev.libclang.override {
            inherit libllvm;
          };
          libraries = super.llvmPackages_11.libraries;
        });
      })
    ];
  };
  pkgsLinux = import <nixpkgs> {
    config = { };
    overlays = [ ];
    system = targetSystem;
  };
in
let
  pkgs = nixpkgs.buildPackages;
  linuxCC = pkgs.wrapCCWith rec {
    cc = pkgs.llvmPackages_11.clang-unwrapped;
    bintools = pkgs.llvmPackages_11.bintools;
    extraPackages = [ pkgsLinux.glibc.static pkgs.llvmPackages_11.libraries.libcxxabi pkgs.llvmPackages_11.libraries.libcxx ];
    extraBuildCommands = ''
      echo "-isystem ${pkgs.llvmPackages_11.clang-unwrapped.lib}/lib/clang/${cc.version}/include" >> $out/nix-support/cc-cflags
      echo "-isystem ${pkgsLinux.glibc.dev}/include" >> $out/nix-support/cc-cflags
      echo "-L ${pkgs.llvmPackages_11.libraries.libcxxabi}/lib" >> $out/nix-support/cc-ldflags
      echo "-L ${pkgsLinux.glibc.static}/lib" >> $out/nix-support/cc-ldflags
      echo "-resource-dir=${cc}/resource-root" >> $out/nix-support/cc-cflags
    '';
  };
in
pkgs.buildEnv (
  let
    cc = linuxCC;
  in
  {
    name = "bazel-${cc.name}-cc";
    paths = [ cc cc.bintools ];
    pathsToLink = [ "/bin" ];
    passthru = {
      inherit (cc) isClang targetPrefix;
      orignalName = cc.name;
    };
  }
)
